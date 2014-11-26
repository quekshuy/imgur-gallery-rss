package main

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/feeds"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	IMGUR_API_HOST_PATH        = "https://api.imgur.com/3"
	IMGUR_GALLERY_API_ENDPOINT = IMGUR_API_HOST_PATH + "/gallery/hot/viral/0.json"
)

type imgurobject struct {
	Link       string `json:"link"`
	ObjectType string `json:"type"`
	Title      string `json:"title"`
	IsAlbum    bool   `json:"is_album"`
	Id         string `json:"id"`
}

func (io *imgurobject) TimesEncounteredKey() string {
	return io.Link + "_count"
}

func (io *imgurobject) IsRepeat() bool {
	isRepeat := false
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println("Error connecting to redis")
	} else {
		newVal, err := redis.Int(conn.Do("INCR", io.TimesEncounteredKey()))
		if err != nil {
			fmt.Println("Error sending command to redis")
		} else {
			if newVal > 1 {
				isRepeat = true
			}
		}
	}

	// finally we should set it to expire in 7 days
	conn.Do("EXPIRE", io.TimesEncounteredKey(), 7*24*60*60)
	return isRepeat
}

type imgurgallerypage struct {
	Data []imgurobject `json:"data"`
}

type imguralbumimageobject struct {
	Link string `json:"link"`
}

type imguralbumimages struct {
	Data []imguralbumimageobject `json:"data"`
}

func (ia *imguralbumimages) Content() string {
	content := ""
	for _, albumImageObj := range ia.Data {
		content += "<img src=\"" + albumImageObj.Link + "\" /><br />"
	}
	return content
}

func (io *imgurobject) AlbumApiEndpointURL() string {
	return IMGUR_API_HOST_PATH + "/album/" + io.Id + "/images"
}

func imgurClientId() string {
	clientId := os.Getenv("IMGUR_CLIENT_ID")
	//fmt.Println("Environment variables CLIENT_ID = ", clientId)
	return clientId
}

func getFeed() *feeds.Feed {
	return &feeds.Feed{
		Title:       "imgur gallery unofficial feed",
		Link:        &feeds.Link{Href: "https://imgur.com/gallery"},
		Description: "My unofficial imgur gallery feed",
		Author:      &feeds.Author{"your name", "your_name@youremail.com"},
		Created:     time.Now(),
	}
}

func sendApiRequest(url string) []byte {
	client := &http.Client{}
	fmt.Println("Requesting ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request to " + url)
	}
	// add the required header
	// https://api.imgur.com/oauth2/addclient
	req.Header.Add("Authorization", "Client-ID "+imgurClientId())
	resp, err := client.Do(req)
	defer resp.Body.Close()
	// JSON body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
	}
	return body
}

func getAlbum(io *imgurobject) *imguralbumimages {
	body := sendApiRequest(io.AlbumApiEndpointURL())
	images := imguralbumimages{}
	json.Unmarshal(body, &images)
	return &images
}

func galleryToFeed(galleryPage *imgurgallerypage) *feeds.Feed {
	// use gorilla feeds to parse into a json RSS feed
	feed := getFeed()
	items := []*feeds.Item{}
	now := time.Now()
	fmt.Println("The number of items = ", len(galleryPage.Data))
	for _, pageItem := range galleryPage.Data {
		var desc string
		if !((&pageItem).IsRepeat()) {
			if pageItem.IsAlbum {
				desc = getAlbum(&pageItem).Content()
			} else {
				desc = "<img src=\"" + pageItem.Link + "\" />"
			}
			items = append(items, &feeds.Item{
				Title:       pageItem.Title,
				Link:        &feeds.Link{Href: pageItem.Link},
				Description: desc,
				Created:     now,
			})
		}
	}
	feed.Items = items
	return feed
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	gallery := imgurgallerypage{}
	body := sendApiRequest(IMGUR_GALLERY_API_ENDPOINT)
	json.Unmarshal(body, &gallery)
	atom, err := galleryToFeed(&gallery).ToAtom()
	if err != nil {
		fmt.Println("Error again lah: ", err)
	} else {
		fmt.Fprintf(w, atom)
	}
}

func main() {
	http.HandleFunc("/imgur/gallery/rss", rssHandler)
	http.ListenAndServe(":8080", nil)
}
