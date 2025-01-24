package models

type IndexRouteProps struct {
	Articles []Article `json:"articles"`
}

type ArticleRouteProps struct {
	Article Article `json:"article"`
}
