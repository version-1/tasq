package entity

const AutoOKLimit = 5

const AutoOKResultPhaseClaim = "claim"

// AutoOKResultInput identifies one pending reply. Repeated acknowledgements are
// idempotent; an old reply cannot update a newly enabled Auto OK cycle.
type AutoOKResultInput struct {
	Generation int64 `json:"generation"`
	// Phase claims a reservation before the runner sends the reply. An empty
	// phase remains the delivery result for backwards compatibility.
	Phase string `json:"phase,omitempty"`
	Error string `json:"error,omitempty"`
}
