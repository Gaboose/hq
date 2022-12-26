package hq

import (
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/participle/v2"
)

var parser = participle.MustBuild[Filter](
	participle.Unquote(),
)

func Exec(queryRaw string, input io.Reader) ([]byte, error) {
	query, err := parser.ParseString("query", queryRaw)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(input)
	if err != nil {
		return nil, err
	}

	val, err := query.Exec(Value{
		selection: doc.Selection,
	})
	if err != nil {
		return nil, err
	}

	return []byte(val.String()), nil
}
