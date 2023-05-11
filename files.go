package main

import (
	"bufio"
	"os"
	"strings"
)

func readLinksFile(filename string) ([]string, error) {
	var links []string

	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		link := scanner.Text()
		link = strings.ReplaceAll(link, "\"", "")
		links = append(links, link)
	}

	file.Close()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
