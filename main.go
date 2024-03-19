package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PostBody struct {
	URL      string `json:"url"`
	Selector string `json:"selector"`
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/rss", rssHandler)
	fmt.Println("Server started at :1337")
	http.ListenAndServe(":1337", addCorsHandler(http.DefaultServeMux))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	url := r.URL.Query().Get("url")
	selector := r.URL.Query().Get("selector")

	if url == "" || selector == "" {
		http.Error(w, "parameters not found: url and selector", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("url not found: %s", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(" failed to read body: %s", err), http.StatusInternalServerError)
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse html: %s", err), http.StatusInternalServerError)
		return
	}

	var innerTexts []string
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		innerText := s.Text()
		innerTexts = append(innerTexts, innerText)
	})

	response := map[string][]string{
		"inner_texts": innerTexts,
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed json: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseJSON)
	if err != nil {
		http.Error(w, fmt.Sprintf("no response: %s", err), http.StatusInternalServerError)
		return
	}
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
    url := r.URL.Query().Get("url")
    selector := r.URL.Query().Get("selector")
    lastpost := r.URL.Query().Get("lastpost")

    if url == "" || selector == "" || lastpost == "" {
        http.Error(w, "Both 'url', 'selector', and 'lastpost' parameters are required", http.StatusBadRequest)
        return
    }

    resp, err := http.Get(url)
    if err != nil {
        
        errorRSS := generateErrorRSS("url not found")
        w.Header().Set("Content-Type", "application/xml")
        w.WriteHeader(http.StatusInternalServerError)
        _, _ = w.Write(errorRSS)
        return
    }
    defer resp.Body.Close()

    bodyBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        
        errorRSS := generateErrorRSS(fmt.Sprintf("no response: %s", err))
        w.Header().Set("Content-Type", "application/xml")
        w.WriteHeader(http.StatusInternalServerError)
        _, _ = w.Write(errorRSS)
        return
    }

    doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
    if err != nil {

        errorRSS := generateErrorRSS(fmt.Sprintf("failed to parse html: %s", err))
        w.Header().Set("Content-Type", "application/xml")
        w.WriteHeader(http.StatusInternalServerError)
        _, _ = w.Write(errorRSS)
        return
    }

    var title string
    doc.Find("title").Each(func(_ int, s *goquery.Selection) {
        title = s.Text()
    })

    var faviconURL string
    doc.Find("link[rel='icon']").Each(func(_ int, s *goquery.Selection) {
        if href, exists := s.Attr("href"); exists {
            faviconURL = href
        }
    })

    var items []Item
    doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
        innerText := s.Text()
        item := Item{
            Title:       title,
            Link:        url,
            Description: innerText,
        }
        items = append(items, item)
    })

    if lastpost == "bottom" {
        for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
            items[i], items[j] = items[j], items[i]
        }
    }

    rss := RSS{
        Version: "2.0",
        Channel: Channel{
            Title:       title,
            Link:        url,
            Description: "ibhub server!",
            Image:       Image{URL: faviconURL},
            Items:       items,
        },
    }

    xmlData, err := xml.MarshalIndent(rss, "", "    ")
    if err != nil {
        
        errorRSS := generateErrorRSS(fmt.Sprintf("failed to generate RSS: %s", err))
        w.Header().Set("Content-Type", "application/xml")
        w.WriteHeader(http.StatusInternalServerError)
        _, _ = w.Write(errorRSS)
        return
    }

    w.Header().Set("Content-Type", "application/xml")
    _, err = w.Write(xmlData)
    if err != nil {
       
        errorRSS := generateErrorRSS(fmt.Sprintf("failed to write XML %s", err))
        w.Header().Set("Content-Type", "application/xml")
        w.WriteHeader(http.StatusInternalServerError)
        _, _ = w.Write(errorRSS)
        return
    }
}

func generateErrorRSS(errorMessage string) []byte {
    rss := RSS{
        Version: "2.0",
        Channel: Channel{
            Title:       "rss error",
            Link:        "",
            Description: errorMessage,
            Image:       Image{URL: ""},
            Items:       []Item{},
        },
    }

    xmlData, err := xml.MarshalIndent(rss, "", "    ")
    if err != nil {
        
        return []byte(errorMessage)
    }

    return xmlData
}

func addCorsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		h.ServeHTTP(w, r)
	})
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Image       Image  `xml:"image"`
	Items       []Item `xml:"item"`
}

type Image struct {
	URL string `xml:"url"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

//example http://localhost:1337/rss?url=https://example.com&selector=.post&lastpost=bottom
