package vecstore

import (
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/components/retriever"

	"github.com/crazyfrankie/voidx/infra/contract/document/progressbar"
)

type IndexerOptions struct {
	PartitionKey   *string
	Partition      *string // Storage sharding map
	IndexingFields []string
	ProgressBar    progressbar.ProgressBar
}

type RetrieverOptions struct {
	MultiMatch   *MultiMatch // Multi-field query
	PartitionKey *string
	Partitions   []string // Query sharding map
}

type MultiMatch struct {
	Fields []string
	Query  string
}

func WithIndexerPartitionKey(key string) indexer.Option {
	return indexer.WrapImplSpecificOptFn(func(o *IndexerOptions) {
		o.PartitionKey = &key
	})
}

func WithPartition(partition string) indexer.Option {
	return indexer.WrapImplSpecificOptFn(func(o *IndexerOptions) {
		o.Partition = &partition
	})
}

func WithIndexingFields(fields []string) indexer.Option {
	return indexer.WrapImplSpecificOptFn(func(o *IndexerOptions) {
		o.IndexingFields = fields
	})
}

func WithProgressBar(progressBar progressbar.ProgressBar) indexer.Option {
	return indexer.WrapImplSpecificOptFn(func(o *IndexerOptions) {
		o.ProgressBar = progressBar
	})
}

func WithMultiMatch(fields []string, query string) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(o *RetrieverOptions) {
		o.MultiMatch = &MultiMatch{
			Fields: fields,
			Query:  query,
		}
	})
}

func WithRetrieverPartitionKey(key string) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(o *RetrieverOptions) {
		o.PartitionKey = &key
	})
}

func WithPartitions(partitions []string) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(o *RetrieverOptions) {
		o.Partitions = partitions
	})
}
