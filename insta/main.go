package main

import (
	// "fmt"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	cache "github.com/patrickmn/go-cache"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var c = cache.New(15*time.Minute, 30*time.Minute) // Lamda lives minutes
var ginLambda *ginadapter.GinLambda

/*Handler - pass control to gin */
func Handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if ginLambda == nil {
		// stdout and stderr are sent to AWS CloudWatch Logs
		log.Printf("Gin cold start")
		r := gin.Default()

		r.GET("/insta/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})
		r.GET("/insta/json", getJSON)

		ginLambda = ginadapter.New(r)
	}

	return ginLambda.Proxy(req)
}

func main() {
	lambda.Start(Handler)
}

func getJSON(c *gin.Context) {
	user := c.DefaultQuery("user", os.Getenv("USERNAME"))
	log.Println("user:", user)
	password := c.DefaultQuery("pwd", os.Getenv("PASSWORD"))
	limit := c.DefaultQuery("limit", "25")
	lmt, _ := strconv.Atoi(limit)
	insta, err := login(user, password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data := instagram(*insta, lmt)
	c.JSON(http.StatusOK, data)
	return
}

/* login
returns goinsta.Instagram object
based on saved JSON object or via new login for user
TODO - better edge cases
*/
func login(user string, password string) (*goinsta.Instagram, error) {
	var insta *goinsta.Instagram
	gc, found := c.Get(user)
	if found {
		log.Println("Found session", user)
		insta = gc.(*goinsta.Instagram)
	} else {
		log.Println("Not found session", user, "Logging with user/password")
		insta = goinsta.New(user, password)
		err := insta.Login()
		if err != nil {
			log.Println(err.Error())
			return insta, err
		}
		c.Set(user, insta, cache.DefaultExpiration)
	}

	return insta, nil
}

/* instagram
returns JSON with images metadata (links, places, likers etc.)
returns <= limit images
processing is slow, takes to long for AWS Proxy timeout
*/
func instagram(insta goinsta.Instagram, limit int) *[]instaImage {
	var Images []instaImage
	media := insta.Account.Feed()
	i := 0
	// Label break (break out of two loops with single break statement)
MediaLoop:
	for media.Next() { // 2-step iteration 1) Going through pages with Next()
		for _, item := range media.Items { // 2) Iterating through items in a page
			i++
			if len(item.Images.Versions) > 0 {
				// Cast image metadata into smaller object
				Image := cast(item)
				// tm := time.Unix(Image.TakenAt, 0)
				// log.Println(i, ":", Image.ID, "-", tm)
				// Append image to array
				// log.Println(Image.ImageVersions2.Candidates[0].URL)
				Images = append(Images, Image)
			}
			if i >= limit {
				break MediaLoop
			} // We only need so many images
		}
	}
	return &Images
}

/* cast - cast struct into JSON, into smaller struct */
func cast(item interface{}) instaImage {
	var Image instaImage
	// create JSON from item
	jsonMedia, jsonErr1 := json.MarshalIndent(item, "    ", "    ")
	if jsonErr1 != nil {
		panic(jsonErr1.Error())
	}
	// Unmarshal JSON into Image
	jsonErr2 := json.Unmarshal(jsonMedia, &Image)
	if jsonErr2 != nil {
		panic(jsonErr2.Error())
	}
	return Image
}

/* instaImage
Instagram Image striped down */
type instaImage struct {
	TakenAt         int64  `json:"taken_at"`
	ID              string `json:"id"`
	DeviceTimestamp int64  `json:"device_timestamp"`
	MediaType       int    `json:"media_type"`
	ClientCacheKey  string `json:"client_cache_key"`
	Caption         struct {
		Text string `json:"text"`
		User struct {
			Username string `json:"username"`
		} `json:"user,omitempty"`
	} `json:"caption"`
	LikeCount      int      `json:"like_count"`
	TopLikers      []string `json:"top_likers,omitempty"`
	ImageVersions2 struct {
		Candidates []struct {
			Width  int    `json:"width"`
			Height int    `json:"height"`
			URL    string `json:"url"`
		} `json:"candidates"`
	} `json:"image_versions2"`
	OriginalWidth  int `json:"original_width"`
	OriginalHeight int `json:"original_height"`
	Location       struct {
		Name      string  `json:"name"`
		City      string  `json:"city"`
		ShortName string  `json:"short_name"`
		Lng       float64 `json:"lng"`
		Lat       float64 `json:"lat"`
	} `json:"location,omitempty"`
}
