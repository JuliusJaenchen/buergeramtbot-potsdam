package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/joho/godotenv/autoload"
)

// This page is the entry link to search for home-regestration
// We need to call it first to get valid cookies
var entryURL = "https://service.berlin.de/terminvereinbarung/termin/all/120686/"

// TODO place current UNIX timestamp in slug. The current one is still hardcoded...
var appointmentURL = "https://service.berlin.de/terminvereinbarung/termin/day/1714428000/"

func main() {
	for {
		poll()
		time.Sleep(time.Minute)
	}
}

func poll() {
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{Jar: jar}

	getCookies(httpClient)

	doc := getAppointmentPage(httpClient)
	bookableAppointmentCount := getBookableAppointmentCount(doc)
	if bookableAppointmentCount > 0 {
		fmt.Print("Success! ")
		req := createTelegramSendMessageRequest("Hey, I've got some free appointments. Go get em'! https://service.berlin.de/terminvereinbarung/termin/all/120686/")
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
		}
	}
	fmt.Printf("Found %v days with open slots (%s)\n", bookableAppointmentCount, time.Now().Format("02.01.2006 15:04:05"))
}

func getCookies(httpClient *http.Client) {
	req := createGetRequest(entryURL)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
}

func getAppointmentPage(httpClient *http.Client) *goquery.Document {
	req := createGetRequest(appointmentURL)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func getBookableAppointmentCount(doc *goquery.Document) int {
	errorMessage := doc.Find(".alert-error")
	if len(errorMessage.Nodes) > 0 {
		log.Fatal("ERROR: Bürgeramt session invalid")
	}

	bookableDataPoints := doc.Find("td.buchbar")
	return len(bookableDataPoints.Nodes)

}

func createGetRequest(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	return req
}
