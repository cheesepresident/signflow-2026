package web

type DocumentView struct {
	ID          string
	Filename    string
	ContentType string
	SizeHuman   string
	FileHash    string
	Status      string
	CreatedAt   string
	IsDraft     bool
}
