package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"example.com/gin/models"

	"github.com/gin-gonic/gin"
	gossr "github.com/natewong1313/go-react-ssr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
)

var APP_ENV string
var bucketName string
var s3Endpoint string
var accessKeyID string
var secretAccessKey string
var s3Region string

func init() {
	bucketName = os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("S3_BUCKET_NAME not set")
	}

	accessKeyID = os.Getenv("S3_ACCESS_KEY")
	secretAccessKey = os.Getenv("S3_SECRET_KEY")
	s3Endpoint = os.Getenv("S3_ENDPOINT")
	s3Region = os.Getenv("S3_REGION")

	if s3Endpoint == "" {
		log.Fatal("S3_ENDPOINT not set")
	}

	log.Printf("S3_BUCKET_NAME: %s", bucketName)
	log.Printf("S3_ACCESS_KEY: %s", accessKeyID)
	log.Printf("S3_SECRET_KEY: %s", secretAccessKey)
	log.Printf("S3_ENDPOINT: %s", s3Endpoint)
	log.Printf("S3_REGION: %s", s3Region)
}

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
		objectKey := "index.html"

		renderedResponse := engine.RenderRoute(gossr.RenderConfig{
			File:     "pages/Home.tsx",
			Title:    "Блог",
			MetaTags: map[string]string{
				"og:title":    "Блог",
				"description": "Главная страница блога",
			},
			Props: &models.IndexRouteProps{
				Articles: articles,
			},
		})
		htmlContent := string(renderedResponse)

		err := uploadHTMLToS3(bucketName, objectKey, htmlContent)
		if err != nil {
			log.Printf("Error uploading to S3: %v", err)
			c.String(http.StatusInternalServerError, "Failed to upload to S3")
			return
		}

		htmlFromS3, err := getHTMLFromS3(bucketName, objectKey)
		if err != nil {
			log.Printf("Error getting HTML from S3: %v", err)
			c.String(http.StatusInternalServerError, "Failed to get HTML from S3")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlFromS3))
	})

	g.GET("/articles/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		currentArticle, err := models.GetArticleBySlug(slug)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		objectKey := fmt.Sprintf("articles/%s.html", slug)

		renderedResponse := engine.RenderRoute(gossr.RenderConfig{
			File:     "pages/ArticlePage.tsx",
			Title:    currentArticle.Title,
			MetaTags: map[string]string{
				"og:title":    currentArticle.Title,
				"description": currentArticle.Description,
			},
			Props: &models.ArticleRouteProps{
				Article: currentArticle,
			},
		})
		htmlContent := string(renderedResponse)

		err = uploadHTMLToS3(bucketName, objectKey, htmlContent)
		if err != nil {
			log.Printf("Error uploading to S3: %v", err)
			c.String(http.StatusInternalServerError, "Failed to upload to S3")
			return
		}

		htmlFromS3, err := getHTMLFromS3(bucketName, objectKey)
		if err != nil {
			log.Printf("Error getting HTML from S3: %v", err)
			c.String(http.StatusInternalServerError, "Failed to get HTML from S3")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlFromS3))
	})
	g.Run(":8080")
}

func uploadHTMLToS3(bucketName, objectKey, htmlContent string) error {
	uploadURL := fmt.Sprintf("%s/%s/%s", s3Endpoint, bucketName, objectKey)

	log.Printf("Uploading to S3: URL=%s", uploadURL)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:               s3Endpoint,
					SigningRegion:     s3Region,
					HostnameImmutable: true,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})),
	)
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg)

	uploader := manager.NewUploader(client)

	body := bytes.NewReader([]byte(htmlContent))

	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   body,
		ContentType: aws.String("text/html"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object, %v", err)
	}

	log.Printf("Successfully uploaded to S3: %s", uploadURL)
	return nil
}

func deleteHTMLFromS3(bucketName, objectKey string) error {
	return nil
}

func getHTMLFromS3(bucketName, objectKey string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:               s3Endpoint,
					SigningRegion:     s3Region,
					HostnameImmutable: true,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})),
	)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg)

	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object, %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body, %v", err)
	}

	return string(body), nil
}
