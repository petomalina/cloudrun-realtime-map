package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/idtoken"
)

func main() {
	ctx := context.Background()

	// pull default credentials from the GCP metadata
	credentials, err := google.FindDefaultCredentials(context.Background(), compute.ComputeScope)
	projectID := credentials.ProjectID
	if err != nil {
		panic(err)
	}

	// create firebase config and specify current project
	conf := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	// get Firestore client
	firestore, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer firestore.Close()

	svc := service{
		h:        resty.New(),
		fs:       firestore,
		iplocUrl: "https://iploc-ygmoaymzvq-ey.a.run.app/locate",
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", svc.pingHandler)

	// Cloud Run passes PORT as an environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

type service struct {
	h  *resty.Client
	fs *firestore.Client

	tokenSource oauth2.TokenSource

	iplocUrl string
}

type iplocResp struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (s *service) pingHandler(c *gin.Context) {
	var err error
	if s.tokenSource == nil {
		s.tokenSource, err = idtoken.NewTokenSource(c.Request.Context(), s.iplocUrl)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("IdToken source could not be created: %w", err))
		}
	}

	token, err := s.tokenSource.Token()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("IdToken could not be retrieved: %w", err))
	}

	resp, err := s.h.R().SetAuthToken(token.AccessToken).Get(s.iplocUrl + "?ip=" + c.ClientIP())
	if err != nil {
		c.String(http.StatusInternalServerError, "An error occurred when calling iploc: "+err.Error())
		return
	}

	var respStruct iplocResp
	err = json.Unmarshal(resp.Body(), &respStruct)
	if err != nil {
		c.String(http.StatusInternalServerError, "An error occurred when unmarhalling iploc resp: "+err.Error())
		return
	}

	h := sha1.New()
	h.Write([]byte(c.ClientIP()))
	bs := h.Sum(nil)

	ipHash := fmt.Sprintf("%x", bs)

	doc := s.fs.Collection("markers").Doc(ipHash)

	color := "#" + c.Query("color")
	if color == "#" {
		color += "005aff"
	}
	doc.Set(
		c.Request.Context(),
		map[string]interface{}{
			"lat":   respStruct.Lat,
			"lng":   respStruct.Lng,
			"color": string(color[:7]),
			"t":     time.Now(),
		},
	)

	c.JSON(200, gin.H{
		"message": "pong",
	})
}
