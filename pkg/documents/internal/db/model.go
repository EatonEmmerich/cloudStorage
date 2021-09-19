package db


type Document struct {
	ID    int64
	Owner   int64
	Path    string
	Version int64
	Size    int64
	MediaType string
	FileName string
}