package entity

const AutoOKLimit = 5

// AutoOKResultInput identifies one pending reply. Repeated acknowledgements are
// idempotent; an old reply cannot update a newly enabled Auto OK cycle.
type AutoOKResultInput struct {
	Generation int64  `json:"generation"`
	Error      string `json:"error,omitempty"`
}
