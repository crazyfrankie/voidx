package file_extractor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/tmc/langchaingo/documentloaders"
)

// FileExtractor handles file extraction and conversion to LangChain documents
type FileExtractor struct {
	minioClient *minio.Client
}

// NewFileExtractor creates a new FileExtractor instance
func NewFileExtractor(minioClient *minio.Client) *FileExtractor {
	return &FileExtractor{
		minioClient: minioClient,
	}
}

// Load loads a file from UploadFile record and returns LangChain documents or text
//func (f *FileExtractor) Load(uploadFile *entity.UploadFile, returnText, isUnstructured bool) (interface{}, error) {
//	// Create a temporary directory
//	tempDir, err := os.MkdirTemp("", "file_extractor_*")
//	if err != nil {
//		return nil, fmt.Errorf("failed to create temp directory: %w", err)
//	}
//	defer os.RemoveAll(tempDir)
//
//	// Create temporary file path
//	filePath := filepath.Join(tempDir, filepath.Base(uploadFile.Key))
//
//	// Download file from object storage
//	if err := f.minioClient.DownloadFile(uploadFile.Key, filePath); err != nil {
//		return nil, fmt.Errorf("failed to download file: %w", err)
//	}
//
//	// Load file from path
//	return LoadFromFile(filePath, returnText, isUnstructured)
//}

// LoadFromURL loads a file from URL and returns LangChain documents or text
func LoadFromURL(url string, returnText bool) (interface{}, error) {
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
func LoadFromFile(filePath string, returnText, isUnstructured bool) (interface{}, error) {
	// Get file extension
	extension := strings.ToLower(filepath.Ext(filePath))

	// Create appropriate loader based on file extension
	var loader documentloaders.Loader
	var err error

	switch extension {
	//case ".xlsx", ".xls":
	//	loader, err = documentloaders.NewExcel(filePath)
	//case ".pdf":
	//	loader = documentloaders.NewPDF(filePath)
	//case ".md", ".markdown":
	//	loader, err = documentloaders.NewText(filePath) // Use text loader for markdown
	//case ".htm", ".html":
	//	loader, err = documentloaders.NewHTML(filePath)
	//case ".csv":
	//	loader, err = documentloaders.NewCSV(filePath)
	//case ".ppt", ".pptx":
	//	loader, err = documentloaders.NewText(filePath) // Use text loader for PowerPoint
	//case ".xml":
	//	loader, err = documentloaders.NewText(filePath) // Use text loader for XML
	//default:
	//	loader, err = documentloaders.NewText(filePath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create document loader: %w", err)
	}

	// Load documents
	docs, err := loader.Load(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load documents: %w", err)
	}

	// Return text or documents based on returnText parameter
	if returnText {
		var texts []string
		for _, doc := range docs {
			texts = append(texts, doc.PageContent)
		}
		return strings.Join(texts, "\n\n"), nil
	}

	return docs, nil
}
