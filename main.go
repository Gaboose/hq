package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/participle/v2"
)

type Suffix struct {
	Iterator bool `@("[" "]")`
}

type Filter struct {
	Array Query  `"[" @@ "]"`
	Find  string `| ("find" "(" @String ")"`
	Html  bool   `| @"html")`

	Suffix Suffix `@@?`
}

type Query struct {
	Filters []Filter `@@ ("|" @@)*`
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

func mainErr() error {
	parser := participle.MustBuild[Query](participle.Unquote())

	queryRaw := flag.Arg(0)
	if queryRaw == "" {
		_, err := io.Copy(os.Stdout, os.Stdin)
		return err
	}

	queryRaw = `[find("table#searchResult th,tr")[] | [find("td")[] | html]]`

	query, err := parser.ParseString("query", queryRaw)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", string(must(json.Marshal(query))))

	doc, err := goquery.NewDocumentFromReader(os.Stdin)
	if err != nil {
		return err
	}

	for _, filter := range query.Filters {
		switch {
		case filter.Find != "":
			doc.Find(filter.Find).Each(func(i int, s *goquery.Selection) {
				html, err := s.Html()
				if err != nil {
					fmt.Println(err)
				}

				fmt.Println(html)
			})
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("Usage: cat some.html | hq [query]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := mainErr(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
