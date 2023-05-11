package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type cfg struct {
	parseOption int
	linksFile   string
}

type Tag interface{}

const (
	img = "img"
	a   = "a"
)

var TagAttr = map[Tag]string{
	img: "src",
	a:   "href",
}

var ParseOptions = map[int]Tag{
	1: a,
	2: img,
}

func getTagAttr(token html.Token, key string) (string, bool) {
	for _, val := range token.Attr {
		if val.Key == key {
			return val.Val, false
		}
	}

	return "", true
}

func ParseHtml(baseUrl string, tag Tag, wg *sync.WaitGroup) error {
	defer wg.Done()
	file, err := ioutil.ReadFile(fmt.Sprintf("./parsed/%s.html", baseUrl))

	if err != nil {
		return err
	}

	token := html.NewTokenizer(strings.NewReader(string(file)))

	var res []string
	var isRequestedTag bool

parser:
	for {
		// get next token
		tt := token.Next()

		switch {
		// if end of line return
		case tt == html.ErrorToken:
			break parser

		// if token is starting token e.g <div>
		case tt == html.StartTagToken:
			t := token.Token()
			isRequestedTag = t.Data == tag

			if isRequestedTag {
				attr, _ := getTagAttr(t, TagAttr[tag])

				// check if its full link or
				// absolute link e.g. "/home"
				if strings.Contains(attr, "http") {
					link := fmt.Sprintf("Link: %s", attr)
					res = append(res, link)
				} else {
					fullLink := fmt.Sprintf("%s/%s", baseUrl, attr)
					link := fmt.Sprintf("Link: %s", fullLink)
					res = append(res, link)
				}
			}

		// if token is text data
		case tt == html.TextToken:
			t := token.Token()

			if isRequestedTag {
				link := fmt.Sprintf("Name: %s\n", strings.TrimSpace(t.Data))
				res = append(res, link)
			}

			isRequestedTag = false
		}
	}

	textFile, _ := os.Create(fmt.Sprintf("./parsed/%s.txt", baseUrl))

	for _, val := range res {
		textFile.WriteString(val + "\n")
	}

	return nil
}

func getUrlDomainName(fullUrl string) (string, error) {
	newUrl, err := url.Parse(fullUrl)

	if err != nil {
		return "", nil
	}

	return newUrl.Hostname(), nil
}

func getHtmlPage(url string, wg *sync.WaitGroup, htmlChan chan string) error {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		fmt.Println(err)
		return err
	}

	client := &http.Client{}

	res, err := client.Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)

	filename, err := getUrlDomainName(url)

	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("./parsed/%s.html", filename)

	htmlFile, err := os.Create(filePath)

	if err != nil {
		return err
	}

	for scanner.Scan() {
		htmlFile.WriteString(scanner.Text())
	}

	htmlChan <- filename

	return err
}

func main() {
	app := &cfg{}

	flag.IntVar(&app.parseOption, "o", 1, "Parse option: 1. Links \n 2. Images")
	flag.StringVar(&app.linksFile, "f", "./links.txt", "File with links to parse")
	flag.Parse()

	urls, err := readLinksFile(app.linksFile)

	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	tag := ParseOptions[app.parseOption]

	if tag == nil {
		fmt.Printf("Error: No such option")
		return
	}

	os.MkdirAll("./parsed", 0755)

	var wg sync.WaitGroup
	htmlChan := make(chan string)
	now := time.Now()

	// get html pages
	for _, v := range urls {
		wg.Add(1)
		go getHtmlPage(v, &wg, htmlChan)
	}

	// wait for all fetching goroutines to finish
	go func() {
		wg.Wait()
		close(htmlChan)
	}()

	// start parsing goroutines
	for v := range htmlChan {
		fmt.Println(v)
		wg.Add(1)
		go ParseHtml(v, tag, &wg)
	}

	wg.Wait()

	fmt.Printf("Successfully parsed data from %s \n", app.linksFile)
	fmt.Printf("Parsed in %g seconds \n", time.Since(now).Seconds())
}
