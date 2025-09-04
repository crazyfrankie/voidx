package file_extractor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/upload"
)

// FileExtractor handles file extraction and conversion to LangChain documents
type FileExtractor struct {
	uploadSvc *upload.Service
}

// NewFileExtractor creates a new FileExtractor instance
func NewFileExtractor(uploadSvc *upload.Service) *FileExtractor {
	return &FileExtractor{
		uploadSvc: uploadSvc,
	}
}

// Load loads a file from UploadFile record and returns LangChain documents or text
func (f *FileExtractor) Load(ctx context.Context, uploadFile *entity.UploadFile, returnText, isUnstructured bool) ([]schema.Document, error) {
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

// LoadFromURL loads a file from URL and returns LangChain documents or text
func LoadFromURL(url string, returnText bool) (any, error) {
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

// LoadFromFile loads a file from local path and returns LangChain documents or text
func LoadFromFile(filePath string, returnText, isUnstructured bool) ([]schema.Document, error) {
	// Get file extension
	extension := strings.ToLower(filepath.Ext(filePath))

	// Create appropriate loader based on file extension
	var loader documentloaders.Loader
	var err error

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	switch extension {
	//case ".xlsx", ".xls":
	//	loader, err = documentloaders.NewExcel(filePath)
	case ".pdf":
		finfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		loader = documentloaders.NewPDF(file, finfo.Size())
	case ".md", ".markdown":
		loader = documentloaders.NewText(file) // Use text loader for markdown
	case ".htm", ".html":
		loader = documentloaders.NewHTML(file)
	case ".csv":
		loader = documentloaders.NewCSV(file)
	case ".ppt", ".pptx":
		loader = documentloaders.NewText(file) // Use text loader for PowerPoint
	case ".xml":
		loader = documentloaders.NewText(file) // Use text loader for XML
	default:
		loader = documentloaders.NewText(file)
	}

	if loader != nil {
		return nil, fmt.Errorf("failed to create document loader: %w", err)
	}

	// Load documents
	docs, err := loader.Load(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load documents: %w", err)
	}

	return docs, nil
}
