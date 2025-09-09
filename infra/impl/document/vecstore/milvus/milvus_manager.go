package milvus

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mentity "github.com/milvus-io/milvus/client/v2/entity"
	mindex "github.com/milvus-io/milvus/client/v2/index"
	client "github.com/milvus-io/milvus/client/v2/milvusclient"

	"github.com/crazyfrankie/voidx/infra/contract/document/vecstore"
	"github.com/crazyfrankie/voidx/infra/contract/embedding"
	"github.com/crazyfrankie/voidx/pkg/lang/ptr"
	"github.com/crazyfrankie/voidx/pkg/lang/sets"
	"github.com/crazyfrankie/voidx/pkg/logs"
	//"github.com/crazyfrankie/voidx/types/errno"
	//"github.com/crazyfrankie/voidx/types/errorx"
)

type ManagerConfig struct {
	Client    *client.Client     // required
	Embedding embedding.Embedder // required

	EnableHybrid *bool              // optional: default Embedding.SupportStatus() == embedding.SupportDenseAndSparse
	DenseIndex   mindex.Index       // optional: default HNSW, M=30, efConstruction=360
	DenseMetric  mentity.MetricType // optional: default IP
	SparseIndex  mindex.Index       // optional: default SPARSE_INVERTED_INDEX, drop_ratio=0.2
	SparseMetric mentity.MetricType // optional: default IP
	ShardNum     int                // optional: default 1
	BatchSize    int                // optional: default 100
}

func NewManager(config *ManagerConfig) (vecstore.Manager, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("[NewManager] milvus client not provided")
	}
	if config.Embedding == nil {
		return nil, fmt.Errorf("[NewManager] milvus embedder not provided")
	}

	enableSparse := config.Embedding.SupportStatus() == embedding.SupportDenseAndSparse
	if config.EnableHybrid == nil {
		config.EnableHybrid = ptr.Of(enableSparse)
	} else if !enableSparse && ptr.From(config.EnableHybrid) {
		logs.Warnf("[NewManager] milvus embedding not support sparse, so hybrid search is disabled.")
		config.EnableHybrid = ptr.Of(false)
	}
	if config.DenseMetric == "" {
		config.DenseMetric = mentity.IP
	}
	if config.DenseIndex == nil {
		config.DenseIndex = mindex.NewHNSWIndex(config.DenseMetric, 30, 360)
	}
	if config.SparseMetric == "" {
		config.SparseMetric = mentity.IP
	}
	if config.SparseIndex == nil {
		config.SparseIndex = mindex.NewSparseInvertedIndex(config.SparseMetric, 0.2)
	}
	if config.ShardNum == 0 {
		config.ShardNum = 1
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}

	return &milvusManager{config: config}, nil
}

type milvusManager struct {
	config *ManagerConfig
}

func (m *milvusManager) Create(ctx context.Context, req *vecstore.CreateRequest) error {
	if err := m.createCollection(ctx, req); err != nil {
		return fmt.Errorf("[Create] create collection failed, %w", err)
	}

	if err := m.createIndexes(ctx, req); err != nil {
		return fmt.Errorf("[Create] create indexes failed, %w", err)
	}

	if exists, err := m.loadCollection(ctx, req.CollectionName); err != nil {
		return fmt.Errorf("[Create] load collection failed, %w", err)
	} else if !exists {
		return fmt.Errorf("[Create] load collection failed, collection=%v does not exist", req.CollectionName)
	}

	return nil
}

func (m *milvusManager) Drop(ctx context.Context, req *vecstore.DropRequest) error {
	return m.config.Client.DropCollection(ctx, client.NewDropCollectionOption(req.CollectionName))
}

func (m *milvusManager) GetType() vecstore.SearchStoreType {
	return vecstore.TypeVectorStore
}

func (m *milvusManager) GetSearchStore(ctx context.Context, collectionName string) (vecstore.SearchStore, error) {
	if exists, err := m.loadCollection(ctx, collectionName); err != nil {
		return nil, err
	} else if !exists {
		//return nil, errorx.New(errno.ErrKnowledgeNonRetryableCode,
		//	errorx.KVf("reason", "[GetSearchStore] collection=%v does not exist", collectionName))
		return nil, errors.New("")
	}

	return &milvusSearchStore{
		config:         m.config,
		collectionName: collectionName,
	}, nil
}

func (m *milvusManager) createCollection(ctx context.Context, req *vecstore.CreateRequest) error {
	if req.CollectionName == "" || len(req.Fields) == 0 {
		return fmt.Errorf("[createCollection] invalid request params")
	}

	cli := m.config.Client
	collectionName := req.CollectionName
	has, err := cli.HasCollection(ctx, client.NewHasCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("[createCollection] HasCollection failed, %w", err)
	}
	if has {
		return nil
	}

	fields, err := m.convertFields(req.Fields)
	if err != nil {
		return err
	}

	opt := client.NewCreateCollectionOption(collectionName, &mentity.Schema{
		CollectionName:     collectionName,
		Description:        fmt.Sprintf("created by coze"),
		AutoID:             false,
		Fields:             fields,
		EnableDynamicField: false,
	}).WithShardNum(int32(m.config.ShardNum))

	for k, v := range req.CollectionMeta {
		opt.WithProperty(k, v)
	}

	if err = cli.CreateCollection(ctx, opt); err != nil {
		return fmt.Errorf("[createCollection] CreateCollection failed, %w", err)
	}

	return nil
}

func (m *milvusManager) createIndexes(ctx context.Context, req *vecstore.CreateRequest) error {
	collectionName := req.CollectionName
	indexes, err := m.config.Client.ListIndexes(ctx, client.NewListIndexOption(req.CollectionName))
	if err != nil {
		if !strings.Contains(err.Error(), "index not found") {
			return fmt.Errorf("[createIndexes] ListIndexes failed, %w", err)
		}
	}

	createdIndexes := sets.FromSlice(indexes)

	var ops []func() error
	for i := range req.Fields {
		f := req.Fields[i]
		if !f.Indexing {
			continue
		}

		ops = append(ops, m.tryCreateIndex(ctx, collectionName, denseFieldName(f.Name), denseIndexName(f.Name), m.config.DenseIndex, createdIndexes))
		if m.config.Embedding.SupportStatus() == embedding.SupportDenseAndSparse {
			ops = append(ops, m.tryCreateIndex(ctx, collectionName, sparseFieldName(f.Name), sparseIndexName(f.Name), m.config.SparseIndex, createdIndexes))
		}
	}

	for _, op := range ops {
		if err := op(); err != nil {
			return fmt.Errorf("[createIndexes] failed, %w", err)
		}
	}

	return nil
}

func (m *milvusManager) tryCreateIndex(ctx context.Context, collectionName, fieldName, indexName string, idx mindex.Index, createdIndexes sets.Set[string]) func() error {
	return func() error {
		if _, found := createdIndexes[indexName]; found {
			logs.CtxInfof(ctx, "[tryCreateIndex] index exists, so skip, collectionName=%s, fieldName=%s, idx=%v, type=%s\n",
				collectionName, fieldName, indexName, idx.IndexType())

			return nil
		}

		cli := m.config.Client

		task, err := cli.CreateIndex(ctx, client.NewCreateIndexOption(collectionName, fieldName, idx).WithIndexName(indexName))
		if err != nil {
			return fmt.Errorf("[tryCreateIndex] CreateIndex failed, %w", err)
		}

		if err = task.Await(ctx); err != nil {
			return fmt.Errorf("[tryCreateIndex] await failed, %w", err)
		}

		logs.CtxInfof(ctx, "[tryCreateIndex] CreateIndex success, collectionName=%s, fieldName=%s, idx=%v, type=%s\n",
			collectionName, fieldName, indexName, idx.IndexType())

		return nil
	}
}

func (m *milvusManager) loadCollection(ctx context.Context, collectionName string) (exists bool, err error) {
	cli := m.config.Client

	stat, err := cli.GetLoadState(ctx, client.NewGetLoadStateOption(collectionName))
	if err != nil {
		return false, fmt.Errorf("[loadCollection] GetLoadState failed, %w", err)
	}

	switch stat.State {
	case mentity.LoadStateNotLoad:
		task, err := cli.LoadCollection(ctx, client.NewLoadCollectionOption(collectionName))
		if err != nil {
			return false, fmt.Errorf("[loadCollection] LoadCollection failed, collection=%v, %w", collectionName, err)
		}
		if err = task.Await(ctx); err != nil {
			return false, fmt.Errorf("[loadCollection] await failed, collection=%v, %w", collectionName, err)
		}
		return true, nil
	case mentity.LoadStateLoaded:
		return true, nil
	case mentity.LoadStateLoading:
		return true, fmt.Errorf("[loadCollection] collection is unloading, retry later, collection=%v", collectionName)
	case mentity.LoadStateUnloading:
		return false, nil
	default:
		return false, fmt.Errorf("[loadCollection] load state unexpected, state=%d", stat)
	}
}

func (m *milvusManager) convertFields(fields []*vecstore.Field) ([]*mentity.Field, error) {
	var foundID, foundCreatorID bool
	resp := make([]*mentity.Field, 0, len(fields))
	for _, f := range fields {
		switch f.Name {
		case vecstore.FieldID:
			foundID = true
		case vecstore.FieldCreatorID:
			foundCreatorID = true
		default:
		}

		if f.Indexing {
			if f.Type != vecstore.FieldTypeText {
				return nil, fmt.Errorf("[convertFields] milvus only support text field indexing, field=%s, type=%d", f.Name, f.Type)
			}
			// Only content is stored when indexing
			if f.Name == vecstore.FieldTextContent {
				resp = append(resp, mentity.NewField().
					WithName(f.Name).
					WithDescription(f.Description).
					WithIsPrimaryKey(f.IsPrimary).
					WithNullable(f.Nullable).
					WithDataType(mentity.FieldTypeVarChar).
					WithMaxLength(65535))
			}
			resp = append(resp, mentity.NewField().
				WithName(denseFieldName(f.Name)).
				WithDataType(mentity.FieldTypeFloatVector).
				WithDim(m.config.Embedding.Dimensions()))
			if m.config.Embedding.SupportStatus() == embedding.SupportDenseAndSparse {
				resp = append(resp, mentity.NewField().
					WithName(sparseFieldName(f.Name)).
					WithDataType(mentity.FieldTypeSparseVector))
			}
		} else {
			mf := mentity.NewField().
				WithName(f.Name).
				WithDescription(f.Description).
				WithIsPrimaryKey(f.IsPrimary).
				WithNullable(f.Nullable)
			typ, err := convertFieldType(f.Type)
			if err != nil {
				return nil, err
			}
			mf.WithDataType(typ)
			if typ == mentity.FieldTypeVarChar {
				mf.WithMaxLength(65535)
			} else if typ == mentity.FieldTypeFloatVector {
				mf.WithDim(m.config.Embedding.Dimensions())
			}
			resp = append(resp, mf)
		}
	}

	if !foundID {
		resp = append(resp, mentity.NewField().
			WithName(vecstore.FieldID).
			WithDataType(mentity.FieldTypeInt64).
			WithIsPrimaryKey(true).
			WithNullable(false))
	}

	if !foundCreatorID {
		resp = append(resp, mentity.NewField().
			WithName(vecstore.FieldCreatorID).
			WithDataType(mentity.FieldTypeInt64).
			WithNullable(false))
	}

	return resp, nil
}

func (m *milvusManager) GetEmbedding() embedding.Embedder {
	return m.config.Embedding
}
