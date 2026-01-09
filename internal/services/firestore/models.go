package firestore

type Database struct {
	Name       string // Short ID: (default)
	ProjectID  string
	Location   string
	Type       string // FIRESTORE_NATIVE / DATASTORE_MODE
	State      string
	CreateTime string
	Uid        string
}
