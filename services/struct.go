package services

type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}
type Response struct {
	ID    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}
