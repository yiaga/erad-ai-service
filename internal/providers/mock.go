package providers

import (
	"context"
	"fmt"

	"github.com/yiaga/erad-ai-service/internal/models"
)

// MockProvider is a test double implementing AIProvider.
type MockProvider struct {
	Name        string
	ShouldError bool
	Result      *ExtractionResult
}

func (m *MockProvider) ProcessDocument(ctx context.Context, input DocumentInput) (*ExtractionResult, error) {
	if m.ShouldError {
		return nil, fmt.Errorf("mock provider error")
	}
	if m.Result != nil {
		return m.Result, nil
	}
	return &ExtractionResult{
		RawText: "mock extracted text for FORM EC 8A Serial Number 0000461",
		ElectionData: &models.ElectionResultData{
			FormEC8ASerialNumber: "0000461",
			ElectionName:         "2026 FCT AREA COUNCIL ELECTIONS MUNICIPAL CHAIRMANSHIP",
			StateCode:            "37",
			State:                "FCT",
			LGACode:              "06",
			LGA:                  "MUNICIPAL",
			RegistrationAreaCode: "03",
			RegistrationArea:     "KABUSA",
			PollingUnitCode:      "139",
			PollingUnit:          "HALAL HOMES STR. PREMIRE ACADEMY",
			PUDelimiter:          "37-06-03-139",
			PresidingOfficer:     "CHUKWUMA-EMEASO, Faith Adazze",
			Date:                 "21/02/2026",
			VotersOnRegister:     747,
			AccreditedVoters:     23,
			BallotIssued:         747,
			UnusedBallot:         723,
			SpoiltBallot:         1,
			RejectedBallot:       0,
			UsedBallot:           24,
			PoliticalParties: models.PoliticalPartyResults{
				"A":   1,
				"ADC": 16,
				"APC": 4,
				"PRP": 1,
				"YPP": 1,
			},
			TotalValidVotes:      23,
			TotalValidVotesWords: "TWENTY THREE",
		},
		ConfidenceScore: 0.99,
		ProviderName:    m.Name,
	}, nil
}

func (m *MockProvider) GetName() string {
	return m.Name
}
