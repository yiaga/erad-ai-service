package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/yiaga/erad-ai-service/internal/models"
	"google.golang.org/api/option"
)

type ResultRefiner struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewResultRefiner(ctx context.Context, apiKey string) (*ResultRefiner, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	
	model := client.GenerativeModel("gemini-2.5-flash")
	// Set JSON response format if supported by SDK version, 
	// otherwise we'll just parse the text.
	model.ResponseMIMEType = "application/json"
	// We rely on the prompt to enforce the schema because ResponseSchema
	// strictly filters out dynamic keys (which we need for political_parties).
	
	return &ResultRefiner{
		client: client,
		model:  model,
	}, nil
}

func (r *ResultRefiner) Refine(ctx context.Context, rawText string) (*models.ElectionResultData, error) {
	prompt := fmt.Sprintf(`
Extract the following fields from the election result sheet OCR text in JSON format.
If a field is not found, use null or 0 as appropriate.
For political parties, extract the number of votes for each party found in the text.

Keys to extract:
- form_ec8a_serial_number
- election_name
- state_code
- state
- lga_code
- lga
- registration_area_code
- registration_area
- polling_unit_code
- polling_unit
- pu_delimiter (StateCode-LGACode-WardCode-PUCode)
- presiding_officer
- date
- voters_on_register
- accredited_voters
- ballot_issued
- unused_ballot
- spoilt_ballot
- rejected_ballot
- used_ballot
- political_parties (JSON object with party name as key and vote count as value)
- total_valid_votes
- total_valid_votes_words

OCR Text:
"""
%s
"""

Return ONLY valid JSON.
`, rawText)

	resp, err := r.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from gemini")
	}

	// Extract JSON from response text (Gemini sometimes adds markdown blocks)
	text := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		text += fmt.Sprintf("%v", part)
	}

	cleanJSON := r.extractJSON(text)
	
	var result models.ElectionResultData
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse gemini json: %w\nRaw text: %s", err, cleanJSON)
	}

	return &result, nil
}

func (r *ResultRefiner) extractJSON(text string) string {
	// Simple cleanup for markdown JSON blocks
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}

func (r *ResultRefiner) Close() {
	if r.client != nil {
		r.client.Close()
	}
}
