package models

type PoliticalPartyResults map[string]int

type ElectionResultData struct {
	FormEC8ASerialNumber string                `json:"form_ec8a_serial_number"`
	ElectionName         string                `json:"election_name"`
	StateCode            string                `json:"state_code"`
	State                string                `json:"state"`
	LGACode              string                `json:"lga_code"`
	LGA                  string                `json:"lga"`
	RegistrationAreaCode string                `json:"registration_area_code"`
	RegistrationArea     string                `json:"registration_area"`
	PollingUnitCode      string                `json:"polling_unit_code"`
	PollingUnit          string                `json:"polling_unit"`
	PUDelimiter          string                `json:"pu_delimiter"`
	PresidingOfficer     string                `json:"presiding_officer"`
	Date                 string                `json:"date"`
	VotersOnRegister     int                   `json:"voters_on_register"`
	AccreditedVoters     int                   `json:"accredited_voters"`
	BallotIssued         int                   `json:"ballot_issued"`
	UnusedBallot         int                   `json:"unused_ballot"`
	SpoiltBallot         int                   `json:"spoilt_ballot"`
	RejectedBallot       int                   `json:"rejected_ballot"`
	UsedBallot           int                   `json:"used_ballot"`
	PoliticalParties     PoliticalPartyResults `json:"political_parties"`
	TotalValidVotes      int                   `json:"total_valid_votes"`
	TotalValidVotesWords string                `json:"total_valid_votes_words"`
}
