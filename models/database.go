package models

import (
 "database/sql"
 "fmt"
 "log"

 _ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
 dbURL := "postgres://user:password@localhost:5432/blog_db?sslmode=disable"

 db, err := sql.Open("postgres", dbURL)
 if err != nil {
  log.Fatalf("Failed to connect to database: %v", err)
 }

 err = db.Ping()
 if err != nil {
  log.Fatalf("Failed to ping database: %v", err)
 }
 fmt.Println("Connected to database successfully!")

 DB = db

 createTable()
}

func createTable() {
 query := `
  CREATE TABLE IF NOT EXISTS articles (
   id SERIAL PRIMARY KEY,
   slug TEXT NOT NULL UNIQUE,
   title TEXT NOT NULL,
   content TEXT NOT NULL,
   description TEXT NOT NULL
  );
 `
 _, err := DB.Exec(query)
 if err != nil {
  log.Fatalf("Failed to create articles table: %v", err)
 }

 fmt.Println("Table created/already exists")
}

func LoadArticles() ([]Article, error) {
 rows, err := DB.Query("SELECT id, slug, title, content, description FROM articles")
 if err != nil {
  return nil, fmt.Errorf("Failed to query articles: %w", err)
 }
 defer rows.Close()

 var articles []Article
 for rows.Next() {
  var article Article
  if err := rows.Scan(&article.ID, &article.Slug, &article.Title, &article.Content, &article.Description); err != nil {
   return nil, fmt.Errorf("Failed to scan article: %w", err)
  }
  articles = append(articles, article)
 }

 if err := rows.Err(); err != nil {
  return nil, fmt.Errorf("Error during rows iteration: %w", err)
 }
 return articles, nil
}

func GetArticleBySlug(slug string) (Article, error) {
 row := DB.QueryRow("SELECT id, slug, title, content, description FROM articles WHERE slug = $1", slug)
 var article Article
 err := row.Scan(&article.ID, &article.Slug, &article.Title, &article.Content, &article.Description)
 if err != nil {
  if err == sql.ErrNoRows {
   return Article{}, fmt.Errorf("Article with slug '%s' not found", slug)
  }
  return Article{}, fmt.Errorf("Failed to scan article: %w", err)
 }
 return article, nil
}

func InsertArticles(articles []Article) error {
 for _, article := range articles {
  _, err := DB.Exec(
   "INSERT INTO articles (slug, title, content, description) VALUES ($1, $2, $3, $4)",
   article.Slug, article.Title, article.Content, article.Description,
  )
  if err != nil {
   return fmt.Errorf("Failed to insert article with slug '%s': %w", article.Slug, err)
  }
 }
 return nil
}

func UpdateArticle(article Article) error {
 _, err := DB.Exec(
  "UPDATE articles SET title = $1, content = $2, description = $3 WHERE slug = $4",
  article.Title, article.Content, article.Description, article.Slug,
 )
 if err != nil {
  return fmt.Errorf("Failed to update article with slug '%s': %w", article.Slug, err)
 }
 return nil
}

func DeleteArticle(slug string) error {
 _, err := DB.Exec("DELETE FROM articles WHERE slug = $1", slug)
 if err != nil {
  return fmt.Errorf("Failed to delete article with slug '%s': %w", slug, err)
 }
 return nil
}
