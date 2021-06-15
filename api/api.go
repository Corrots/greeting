package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// Create a http client
func createClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	return &http.Client{Transport: transport}
}

func Fetch(url string) io.Reader {
	client := createClient()
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("Fetch from url: %s err: %s", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("StatusCode err: %d %s", resp.StatusCode, resp.Status)
	}
	return resp.Body
}

func FetchHTML(url string) (doc *goquery.Document) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatalf("Fetch html error: %s", err)
	}
	return
}

// Get Weather data
func GetWeather(local string) Weather {
	url := "https://tianqi.moji.com/weather/china/" + local
	doc := FetchHTML(url)
	wrap := doc.Find(".wea_info")
	city := wrap.Find(".search_default").First().Text()
	temp := wrap.Find(".wea_weather em").Text()
	weather := wrap.Find(".wea_weather b").Text()
	air := wrap.Find(".wea_alert .level").Next().Text()
	humidity := wrap.Find(".wea_about span").Text()
	wind := wrap.Find(".wea_about em").Text()
	note := wrap.Find(".wea_tips em").Text()
	limit := ""
	if limitDesc := wrap.Find(".wea_about b").Text(); limitDesc != "" {
		desc := []rune(limitDesc)
		if len(desc) <= 4 {
			limit = string(desc)
		} else {
			limit = string(desc[4:])
		}
	}

	return Weather{
		City:     city,
		Temp:     temp,
		Weather:  weather,
		Air:      air,
		Humidity: humidity,
		Wind:     wind,
		Limit:    limit,
		Note:     note,
	}
}

// Get One data
func GetOne() One {
	url := "http://www.vgtime.com/"
	doc := FetchHTML(url)
	wrap := doc.Find(".foc_list ul li").First()
	sentence, exists := wrap.Find(".img_box a").Attr("title")
	if !exists {
		sentence = wrap.Find(".info_box>a>h2").Text()
	}
	date := wrap.Find(".info_box .fot_box .time_box").Children().Nodes[0].NextSibling.Data
	imageUrl, exists := wrap.Find(".img_box img").Attr("src")

	return One{
		Date:     date,
		ImgURL:   imageUrl,
		Sentence: sentence,
	}
}

// Get English data
func GetEnglish() English {
	url := "http://dict.eudic.net/home/dailysentence"
	doc := FetchHTML(url)
	wrap := doc.Find("#getLang .head-img")
	imageUrl, _ := wrap.Find(".himg").Attr("src")
	sentence := wrap.Find(".sentence .sect_en").Text()

	return English{
		ImgURL:   imageUrl,
		Sentence: sentence,
	}
}

// Get Poem data
func GetPoem() Poem {
	url := "https://v2.jinrishici.com/one.json"
	result := Fetch(url)

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(result); err != nil {
		log.Fatalf("convert io reader to []byte err: %s", err)
	}
	resBytes := buf.Bytes()

	var resJSON PoemRes
	if err := json.Unmarshal(resBytes, &resJSON); err != nil {
		log.Fatalf("Unmarshal json failed: %s", err)
	}
	if resJSON.Status != "success" {
		log.Fatalf("Get Poem status: %s", resJSON.Status)
	}

	return resJSON.Data.Origin
}

// Get Wallpaper from Bing
func GetWallpaper() Wallpaper {
	url := "https://www.bing.com"
	doc := FetchHTML(url + "?mkt=zh-CN")
	imageUrl, _ := doc.Find("#bgLink").Attr("href")
	title, _ := doc.Find("#sh_cp").Attr("title")
	return Wallpaper{
		ImgURL: url + imageUrl,
		Title:  title,
	}
}

// GetTrivia data
func GetTrivia() Trivia {
	url := "http://www.lengdou.net/random"
	doc := FetchHTML(url)
	imageUrl, _ := doc.Find(".topic-img img").Attr("src")
	description := doc.Find(".topic-content").Text()

	return Trivia{
		ImgURL:      imageUrl,
		Description: description,
	}
}
