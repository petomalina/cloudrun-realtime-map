package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

var cache map[string]ipstackResponse

func main() {
	cache = make(map[string]ipstackResponse)

	svc := service{
		h: resty.New(),
	}

	r := gin.Default()
	r.GET("/locate", svc.locateHandler)

	// Cloud Run passes PORT as an environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

type service struct {
	h *resty.Client
}

type ipstackResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (s *service) locateHandler(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		c.String(http.StatusOK, "It's ok, we all make mistakes. But PLEASE, read the %$@!ing docs")
		return
	}

	if _, ok := cache[ip]; !ok {
		resp, err := s.h.R().Get("http://api.ipstack.com/" + ip + "?access_key=" + os.Getenv("IPSTACK_KEY"))
		if err != nil {
			c.String(http.StatusInternalServerError, "An error occurred when calling ipstack: "+err.Error())
			return
		}

		var respStruct ipstackResponse
		err = json.Unmarshal(resp.Body(), &respStruct)
		if err != nil {
			c.String(http.StatusInternalServerError, "An error occurred when unmarshalling ipstack resp: "+err.Error())
			return
		}

		cache[ip] = respStruct
	}

	loc := cache[ip]

	c.JSON(http.StatusOK, gin.H{
		"lat": loc.Latitude,
		"lng": loc.Longitude,
	})
}
