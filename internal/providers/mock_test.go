package providers

import (
	"context"
	"testing"
)

func TestMockProvider_Success(t *testing.T) {
	provider := &MockProvider{Name: "mock"}
	result, err := provider.ProcessDocument(context.Background(), DocumentInput{
		LocalPath: "/fake/path/image.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProviderName != "mock" {
		t.Errorf("expected provider name 'mock', got '%s'", result.ProviderName)
	}
	if result.ConfidenceScore != 0.99 {
		t.Errorf("expected confidence 0.99, got %f", result.ConfidenceScore)
	}
}

func TestMockProvider_Error(t *testing.T) {
	provider := &MockProvider{Name: "mock", ShouldError: true}
	_, err := provider.ProcessDocument(context.Background(), DocumentInput{
		LocalPath: "/fake/path/image.jpg",
	})
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}
