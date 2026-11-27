package web

type IntegrityView struct {
	Matches     bool
	StoredHash  string // recorded at upload
	CurrentHash string // recomputed now
}

type SignerView struct {
	Email         string
	Status        string
	IsSigned      bool
	SignedName    string
	SignedAt      string
	SignedDocHash string
	// HashMatchesNow is true when this signature's hash still equals the file's
	// current stored hash (i.e. the file wasn't changed after they signed).
	HashMatchesNow bool
}

// SignerRoster is the owner-facing summary of who has and hasn't signed.
type SignerRoster struct {
	Signers []SignerView
	Total   int
	Signed  int
}
type AuditView struct {
	Time       string
	ActorLabel string
	ActorType  string // user | signer | system | anonymous
	Message    string
	EventType  string
	// Security marks a row that deserves visual emphasis (bad token / tamper).
	Security bool
}
