package model

type CurrencyService interface {
	ConvertCurrency(params *ConvertCurrencyParams) (map[string]float64, error)
}

type NoteService interface {
	AddNote(note *Note) error
	GetAllNotes() []Note
	GetNoteByHeader(header string) (*Note, error)
	UpdateNote(note *Note) error
	DeleteNote(header string) error
}

type Service interface {
	NoteService() NoteService
	CurrencyService() CurrencyService
}
