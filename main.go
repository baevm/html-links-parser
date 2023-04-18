package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
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

func ParseHtml(text *string, baseUrl string, tag Tag) []string {
	token := html.NewTokenizer(strings.NewReader(*text))

	var res []string
	var isRequestedTag bool

	for {
		// get next token
		tt := token.Next()

		switch {
		// if end of line return
		case tt == html.ErrorToken:
			return res

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
					fullLink := baseUrl + attr
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
}

func getHtmlPage(url string) (string, error) {
	res, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)

	htmlFile, _ := os.Create("./parsed/Main.html")

	for scanner.Scan() {
		htmlFile.WriteString(scanner.Text())
	}

	text, err := readHtmlFromFile("./parsed/Main.html")

	return text, err
}

func main() {
	var url string
	var parseOption int
	var tag Tag

	fmt.Println("Enter url you want to parse: ")
	fmt.Scan(&url)

	fmt.Printf("Select what you want to parse? \n 1. Links \n 2. Images \n")
	fmt.Scan(&parseOption)

	tag = ParseOptions[parseOption]

	if tag == nil {
		fmt.Printf("Error: No such option")
		return
	}

	os.MkdirAll("./parsed", 0755)

	now := time.Now()
	text, err := getHtmlPage(url)

	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	data := ParseHtml(&text, url, tag)

	textFile, _ := os.Create("./parsed/data.txt")

	for _, val := range data {
		textFile.WriteString(val + "\n")
	}

	fmt.Printf("Successfully parsed data from %s\n", url)
	fmt.Printf("Parsed in %g seconds \n", time.Now().Sub(now).Seconds())
}
