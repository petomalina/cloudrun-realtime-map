package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
)

func main() {
	ctx := context.Background()

	// pull default credentials from the GCP metadata
	credentials, err := google.FindDefaultCredentials(context.Background(), compute.ComputeScope)
	if err != nil {
		panic(err)
	}

	projectID := credentials.ProjectID

	// create firebase config and specify current project
	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	// get Firestore client
	fs, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer fs.Close()

	svc := service{
		fs: fs,
	}

	r := gin.Default()
	r.GET("/cleanup", svc.cleanupHandler)

	// Cloud Run passes PORT as an environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

type service struct {
	fs *firestore.Client
}

func (s *service) cleanupHandler(c *gin.Context) {
	col := s.fs.Collection("markers")

	// get all documents that have been added before time.Now - 40 second
	iter := col.Where("t", "<", time.Now().Add(-time.Second*40)).Documents(c.Request.Context())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("An error occured when iterating through cleanup: %w", err))
			return
		}

		doc.Ref.Delete(c.Request.Context())
	}

	c.JSON(200, gin.H{
		"message": "done",
	})
}
