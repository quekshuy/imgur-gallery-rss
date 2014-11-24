package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/feeds"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const IMGUR_GALLERY_API_ENDPOINT = "https://api.imgur.com/3/gallery/hot/viral/0.json"

type imgurobject struct {
	Link       string `json:"link"`
	ObjectType string `json:"type"`
	Title      string `json:"title"`
}

type imgurgallerypage struct {
	Data []imgurobject `json:"data"`
}

func imgurClientId() string {
	clientId := os.Getenv("IMGUR_CLIENT_ID")
	fmt.Println("Environment variables CLIENT_ID = ", clientId)
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

func galleryToFeed(galleryPage *imgurgallerypage) *feeds.Feed {
	// use gorilla feeds to parse into a json RSS feed
	feed := getFeed()
	items := []*feeds.Item{}
	now := time.Now()
	fmt.Println("The number of items = ", len(galleryPage.Data))
	for _, pageItem := range galleryPage.Data {
		items = append(items, &feeds.Item{
			Title:       pageItem.Title,
			Link:        &feeds.Link{Href: pageItem.Link},
			Description: "<img src=\"" + pageItem.Link + "\" />",
			Created:     now,
		})
	}
	feed.Items = items
	return feed
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	// query imgur to get the client ID
	client := &http.Client{}
	req, err := http.NewRequest("GET", IMGUR_GALLERY_API_ENDPOINT, nil)
	if err != nil {
		fmt.Println("Error ah : ", err)
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
	gallery := imgurgallerypage{}
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
