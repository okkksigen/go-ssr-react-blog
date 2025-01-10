package models

type Article struct {
 Slug        string `json:"slug"`
 Title       string `json:"title"`
 Content     string `json:"content"`
 Description string `json:"description"`
}
