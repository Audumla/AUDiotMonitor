package capabilities

// Requirement describes a runtime dependency required by an adapter.
type Requirement struct {
	Name        string
	Description string
	Optional    bool
}

// Provider is implemented by adapters that expose runtime dependency requirements.
type Provider interface {
	Requirements() []Requirement
}
