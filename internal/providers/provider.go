package providers

import (
	"context"
	"github.com/yiaga/erad-ai-service/internal/models"
)

type DocumentInput struct {
	LocalPath string
	MimeType  string
}

type ExtractionResult struct {
	RawText         string
	StructuredData  interface{}
	ElectionData    *models.ElectionResultData
	ConfidenceScore float64
	ProviderName    string
}

type AIProvider interface {
	ProcessDocument(ctx context.Context, input DocumentInput) (*ExtractionResult, error)
	GetName() string
}
