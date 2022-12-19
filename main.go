package main

import (
	"errors"
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
	Array *Query `"[" @@ "]"`
	Find  string `| ("find" "(" @String ")"`
	Html  bool   `| @"html"`
	Text  bool   `| @"text")`

	Suffix Suffix `@@?`
}

func (f *Filter) Exec(val Value) (Value, error) {
	fmt.Printf("Filter EXEC %+v %+v\n", *f, val)
	// fmt.Printf("DEBUG %+v %+v\n", *f, val)

	var err error

	switch {
	case f.Array != nil:
		val, err = f.Array.Exec(val)
		if err != nil {
			return Value{}, err
		}

		if val.isIterator {
			val.isIterator = false
		} else {
			val = Value{
				array: []Value{val},
			}
		}

		fmt.Printf("DEBUG ARRAY %+v\n", val)
	case f.Find != "":
		val = Value{
			selection: val.selection.Find(f.Find),
		}
	case f.Html:
		s, err := val.selection.Html()
		if err != nil {
			return Value{}, err
		}

		val = Value{
			str: s,
		}
	case f.Text:
		val = Value{
			str: val.selection.Text(),
		}
	}

	if f.Suffix.Iterator {
		val, err = val.iterator()
		if err != nil {
			return Value{}, err
		}
	}

	fmt.Printf("Filter RETURN %+v\n", val)

	return val, nil
}

type Query struct {
	Filters []Filter `@@ ("|" @@)*`
}

func (q *Query) Exec(val Value) (Value, error) {
	fmt.Printf("Query EXEC %+v %+v\n", *q, val)
	var err error

	for _, f := range q.Filters {
		var nextVal Value

		if val.isIterator {
			nextVal := Value{
				isIterator: true,
			}

			for _, item := range val.array {
				nextItem, err := f.Exec(item)
				if err != nil {
					return Value{}, err
				}

				nextVal.array = append(nextVal.array, nextItem)
			}
		} else {
			nextVal, err = f.Exec(val)
			if err != nil {
				return Value{}, err
			}
		}

		val = nextVal
	}

	fmt.Printf("Query RETURN %+v\n", val)

	return val, nil
}

type Value struct {
	array      []Value
	isIterator bool

	selection *goquery.Selection
	str       string
}

func (v *Value) children() ([]Value, error) {
	switch {
	case len(v.array) > 0:
		return v.array, nil
	case v.selection != nil:
		ret := []Value{}
		v.selection.Each(func(_ int, s *goquery.Selection) {
			ret = append(ret, Value{
				selection: s,
			})
		})
		return ret, nil
	case v.str != "":
		return nil, errors.New("tried to iterate string")
	default:
		return nil, errors.New("tried to iterate empty value")
	}
}

func (v *Value) iterator() (Value, error) {
	if v.isIterator {
		nextVal := Value{}

		for _, item := range v.array {
			if item.isIterator {
				panic("only top value can be an iterator")
			}

			children, err := item.children()
			if err != nil {
				return Value{}, err
			}

			nextVal.array = append(nextVal.array, children...)
		}

		return nextVal, nil
	}

	children, err := v.children()
	if err != nil {
		return Value{}, err
	}

	return Value{
		array:      children,
		isIterator: true,
	}, nil
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

	queryRaw = `[find("table#searchResult th,tr")[] | [find("td")[] | text]]`
	// queryRaw = `[find("table#searchResult th,tr")]`

	query, err := parser.ParseString("query", queryRaw)
	if err != nil {
		return err
	}

	// fmt.Printf("%+v\n", string(must(json.Marshal(query))))

	doc, err := goquery.NewDocumentFromReader(os.Stdin)
	if err != nil {
		return err
	}

	val, err := query.Exec(Value{
		selection: doc.Selection,
	})
	if err != nil {
		return err
	}

	// _ = val
	fmt.Printf("%+v\n", val)

	// for _, filter := range query.Filters {
	// 	switch {
	// 	case filter.Find != "":
	// 		doc.Find(filter.Find).Each(func(i int, s *goquery.Selection) {
	// 			html, err := s.Html()
	// 			if err != nil {
	// 				fmt.Println(err)
	// 			}

	// 			fmt.Println(html)
	// 		})
	// 	}
	// }

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
