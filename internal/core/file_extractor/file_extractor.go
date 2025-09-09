package file_extractor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/upload"
)

// FileExtractor handles file extraction and conversion to LangChain documents
type FileExtractor struct {
	uploadSvc  *upload.Service
	fileLoader *file.FileLoader
}

// NewFileExtractor creates a new FileExtractor instance
func NewFileExtractor(uploadSvc *upload.Service) (*FileExtractor, error) {
	// Create eino file loader
	fileLoader, err := file.NewFileLoader(context.Background(), &file.FileLoaderConfig{
		UseNameAsID: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file loader: %w", err)
	}

	return &FileExtractor{
		uploadSvc:  uploadSvc,
		fileLoader: fileLoader,
	}, nil
}

// Load loads a file from UploadFile record and returns eino documents
func (f *FileExtractor) Load(ctx context.Context, uploadFile *entity.UploadFile, returnText, isUnstructured bool) ([]*schema.Document, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "file_extractor_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create temporary file path
	filePath := filepath.Join(tempDir, filepath.Base(uploadFile.Key))

	// Download file from object storage
	if err := f.uploadSvc.DownloadFile(ctx, uploadFile.Key, filePath); err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	// Load file from path
	return LoadFromFile(filePath, returnText, isUnstructured)
}

// LoadFromURL loads a file from URL and returns eino documents
func LoadFromURL(url string, returnText bool) ([]*schema.Document, error) {
	// Download file from URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from URL: %w", err)
	}
	defer resp.Body.Close()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "file_extractor_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create temporary file
	filePath := filepath.Join(tempDir, filepath.Base(url))
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// Copy downloaded content to file
	if _, err := io.Copy(file, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	return LoadFromFile(filePath, returnText, true)
}

// LoadFromFile loads a file from local path and returns eino documents
func LoadFromFile(filePath string, returnText, isUnstructured bool) ([]*schema.Document, error) {
	// Create eino file loader
	fileLoader, err := file.NewFileLoader(context.Background(), &file.FileLoaderConfig{
		UseNameAsID: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file loader: %w", err)
	}

	// Create document source
	src := document.Source{
		URI: filePath,
	}

	// Load documents using eino
	docs, err := fileLoader.Load(context.Background(), src)
	if err != nil {
		return nil, fmt.Errorf("failed to load documents: %w", err)
	}

	return docs, nil
}
