package main

import (
	"bufio"
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

func readHtmlFromFile(filename string) (string, error) {
	bs, err := os.ReadFile(filename)

	if err != nil {
		return "", err
	}

	return string(bs), err
}

func getTagAttr(token html.Token, key string) (string, bool) {
	for _, val := range token.Attr {
		if val.Key == key {
			return val.Val, false
		}
	}

	return "", true
}

func ParseHtml(baseUrl string, tag Tag) error {
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
				link := fmt.Sprintf("Name: %s\n", t.Data)
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

func getHtmlPage(url string, wg *sync.WaitGroup, htmlChan chan string) (string, error) {
	defer wg.Done()

	res, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)

	filename, _ := getUrlDomainName(url)

	filePath := fmt.Sprintf("./parsed/%s.html", filename)

	htmlFile, _ := os.Create(filePath)

	for scanner.Scan() {
		htmlFile.WriteString(scanner.Text())
	}

	htmlChan <- filename

	return filePath, err
}

func main() {
	var parseOption int
	var tag Tag
	urls := []string{
		"https://habr.com/ru/top/weekly/",
		"https://roadmap.sh/frontend",
		"https://stackoverflow.com/",
		"https://vc.ru",
		"https://dtf.ru",
		"https://gobyexample.com/",
	}

	// fmt.Println("Enter urls you want to parse: ")

	// scanner := bufio.NewScanner(os.Stdin)
	// if scanner.Scan() {
	// 	line := scanner.Text()
	// 	urls = append(urls, line)
	// }

	fmt.Printf("Select what you want to parse? \n 1. Links \n 2. Images \n")
	fmt.Scan(&parseOption)

	tag = ParseOptions[parseOption]

	if tag == nil {
		fmt.Printf("Error: No such option")
		return
	}

	os.MkdirAll("./parsed", 0755)

	now := time.Now()
	htmlChan := make(chan string)
	var wg sync.WaitGroup

	// get html pages
	for _, v := range urls {
		wg.Add(1)
		go getHtmlPage(v, &wg, htmlChan)
	}

	go func() {
		wg.Wait()
		close(htmlChan)
	}()

	for v := range htmlChan {
		fmt.Println(v)
		ParseHtml(v, tag)
	}

	fmt.Printf("Successfully parsed data from %s\n", urls)
	fmt.Printf("Parsed in %g seconds \n", time.Now().Sub(now).Seconds())
}
