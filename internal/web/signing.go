package web

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
