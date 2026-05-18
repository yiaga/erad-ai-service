package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type AzureProvider struct {
	apiKey   string
	endpoint string
}

func NewAzureProvider(apiKey, endpoint string) *AzureProvider {
	return &AzureProvider{
		apiKey:   apiKey,
		endpoint: endpoint,
	}
}

func (p *AzureProvider) ProcessDocument(ctx context.Context, input DocumentInput) (*ExtractionResult, error) {
	file, err := os.Open(input.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 1. Submit Document for Analysis
	// We use the 'prebuilt-layout' model which is excellent for election result sheets (tables)
	url := fmt.Sprintf("%s/formrecognizer/documentModels/prebuilt-layout:analyze?api-version=2023-07-31", p.endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, file)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", p.apiKey)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("azure api returned error (%d): %s", resp.StatusCode, string(body))
	}

	operationLocation := resp.Header.Get("Operation-Location")
	if operationLocation == "" {
		return nil, fmt.Errorf("missing Operation-Location header from Azure")
	}

	// 2. Poll for Results
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			pollReq, err := http.NewRequestWithContext(ctx, "GET", operationLocation, nil)
			if err != nil {
				return nil, err
			}
			pollReq.Header.Set("Ocp-Apim-Subscription-Key", p.apiKey)

			pollResp, err := http.DefaultClient.Do(pollReq)
			if err != nil {
				return nil, err
			}
			defer pollResp.Body.Close()

			var result struct {
				Status        string                 `json:"status"`
				AnalyzeResult map[string]interface{} `json:"analyzeResult"`
				Error         *struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(pollResp.Body).Decode(&result); err != nil {
				return nil, err
			}

			if result.Status == "succeeded" {
				return &ExtractionResult{
					RawText:         fmt.Sprintf("%v", result.AnalyzeResult["content"]),
					StructuredData:  result.AnalyzeResult,
					ConfidenceScore: 0.95, // Aggregate confidence
					ProviderName:    p.GetName(),
				}, nil
			}

			if result.Status == "failed" {
				errMsg := "unknown error"
				if result.Error != nil {
					errMsg = result.Error.Message
				}
				return nil, fmt.Errorf("azure analysis failed: %s", errMsg)
			}
		}
	}
}

func (p *AzureProvider) GetName() string {
	return "azure"
}
