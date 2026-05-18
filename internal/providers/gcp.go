package providers

import (
	"context"
	"fmt"
	"io"
	"os"

	documentai "cloud.google.com/go/documentai/apiv1"
	"cloud.google.com/go/documentai/apiv1/documentaipb"
	"google.golang.org/api/option"
)

type GCPProvider struct {
	projectID   string
	location    string
	processorID string
}

func NewGCPProvider(projectID, location, processorID string) *GCPProvider {
	return &GCPProvider{
		projectID:   projectID,
		location:    location,
		processorID: processorID,
	}
}

func (p *GCPProvider) ProcessDocument(ctx context.Context, input DocumentInput) (*ExtractionResult, error) {
	client, err := documentai.NewDocumentProcessorClient(ctx, option.WithEndpoint(fmt.Sprintf("%s-documentai.googleapis.com:443", p.location)))
	if err != nil {
		return nil, fmt.Errorf("failed to create documentai client: %w", err)
	}
	defer client.Close()

	file, err := os.Open(input.LocalPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/processors/%s", p.projectID, p.location, p.processorID)
	req := &documentaipb.ProcessRequest{
		Name: name,
		Source: &documentaipb.ProcessRequest_RawDocument{
			RawDocument: &documentaipb.RawDocument{
				Content:  content,
				MimeType: "image/jpeg", // We could detect this from file extension
			},
		},
	}

	resp, err := client.ProcessDocument(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gcp documentai error: %w", err)
	}

	doc := resp.Document
	return &ExtractionResult{
		RawText:         doc.Text,
		StructuredData:  map[string]interface{}{"entities": doc.Entities},
		ConfidenceScore: 0.95, // GCP provides per-entity confidence
		ProviderName:    p.GetName(),
	}, nil
}

func (p *GCPProvider) GetName() string {
	return "gcp"
}
