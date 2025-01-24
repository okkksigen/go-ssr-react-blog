package models

type Article struct {
	ID          int    `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Description string `json:"description"`
}
