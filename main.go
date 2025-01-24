package main

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "os"

  "example.com/gin/models"

  "github.com/gin-gonic/gin"
  gossr "github.com/natewong1313/go-react-ssr"
)

var APP_ENV string

func loadArticlesFromJSON() ([]models.Article, error) {
  jsonPath := "./data/articles.json"

  file, err := os.ReadFile(jsonPath)
  if err != nil {
    return nil, fmt.Errorf("Failed to read articles.json: %w", err)
  }

  var articles []models.Article
  if err := json.Unmarshal(file, &articles); err != nil {
    return nil, fmt.Errorf("Failed to unmarshal articles.json: %w", err)
  }

  return articles, nil
}

func main() {
  models.InitDB()

  articles, err := models.LoadArticles()

  if err != nil {
    log.Fatalf("Failed to load articles: %v", err)
  }

  if len(articles) == 0 {
    fmt.Println("Database is empty, attempting to load from JSON...")
    articlesFromJSON, err := loadArticlesFromJSON()
    if err != nil {
      log.Fatalf("Failed to load articles from json: %v", err)
    }
    fmt.Println("JSON data loaded successfully. Number of articles:", len(articlesFromJSON))
    err = models.InsertArticles(articlesFromJSON)
    if err != nil {
      log.Fatalf("Failed to insert articles into database: %v", err)
    }
    fmt.Println("Data inserted successfully.")

    articles, err = models.LoadArticles()
    if err != nil {
      log.Fatalf("Failed to load articles from database after insert: %v", err)
    }
  }

  fmt.Printf("Number of articles loaded from DB: %d\n", len(articles))

  g := gin.Default()
  g.StaticFile("favicon.ico", "./frontend/public/favicon.ico")
  g.Static("/assets", "./frontend/public")

  engine, err := gossr.New(gossr.Config{
    AppEnv:             APP_ENV,
    AssetRoute:         "/assets",
    FrontendDir:        "./frontend/src",
    GeneratedTypesPath: "./frontend/src/generated.d.ts",
    TailwindConfigPath: "./frontend/tailwind.config.js",
    LayoutCSSFilePath:  "Main.css",
    PropsStructsPath:   "./models/props.go",
  })
  if err != nil {
    log.Fatal("Failed to init go-react-ssr")
  }

  g.GET("/", func(c *gin.Context) {
    renderedResponse := engine.RenderRoute(gossr.RenderConfig{
      File: "pages/Home.tsx",
      Title: "Блог",
      MetaTags: map[string]string{
        "og:title":    "Блог",
        "description": "Главная страница блога",
      },
      Props: &models.IndexRouteProps{
        Articles: articles,
      },
    })
    c.Writer.Write(renderedResponse)
  })

  g.GET("/articles/:slug", func(c *gin.Context) {
    slug := c.Param("slug")
    currentArticle, err := models.GetArticleBySlug(slug)
    if err != nil {
      c.Status(http.StatusNotFound)
      return
    }

    renderedResponse := engine.RenderRoute(gossr.RenderConfig{
      File: "pages/ArticlePage.tsx",
      Title: currentArticle.Title,
      MetaTags: map[string]string{
        "og:title":    currentArticle.Title,
        "description": currentArticle.Description,
      },
      Props: &models.ArticleRouteProps{
        Article: currentArticle,
      },
    })
    c.Writer.Write(renderedResponse)
  })

  g.Run()
}
