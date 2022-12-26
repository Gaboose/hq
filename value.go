package hq

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/PuerkitoBio/goquery"
)

type ValueKind int

const (
	ValueKindEmpty = iota
	ValueKindArray
	ValueKindIterator
	ValueKindSelection
	ValueKindString
)

type Value struct {
	array    []Value
	iterator []Value

	selection *goquery.Selection
	str       string

	kind ValueKind
}

func (v Value) Array() Value {
	if v.kind == ValueKindIterator {
		return Value{
			kind:  ValueKindArray,
			array: v.iterator,
		}
	}

	return Value{
		kind:  ValueKindArray,
		array: []Value{v},
	}
}

func (v Value) Find(s string) Value {
	return Value{
		kind:      ValueKindSelection,
		selection: v.selection.Find(s),
	}
}

func (v Value) Attr(attrName string) Value {
	s, ok := v.selection.Attr(attrName)
	if !ok {
		return Value{}
	}

	return Value{
		kind: ValueKindString,
		str:  s,
	}
}

func (v Value) Html() (Value, error) {
	s, err := v.selection.Html()
	return Value{
		kind: ValueKindString,
		str:  s,
	}, err
}

func (v Value) Text() Value {
	return Value{
		kind: ValueKindString,
		str:  v.selection.Text(),
	}
}

func (v *Value) Index(i int) (Value, error) {
	var n int

	switch v.kind {
	case ValueKindArray:
		n = len(v.array)
	case ValueKindSelection:
		n = v.selection.Length()
	case ValueKindString:
		return Value{}, errors.New("tried to index string")
	default:
		panic("unexpected value kind")
	}

	if i < 0 {
		i += n
	}

	if i < 0 || i >= n {
		return Value{}, nil
	}

	switch v.kind {
	case ValueKindArray:
		return v.array[i], nil
	case ValueKindSelection:
		return Value{
			kind:      ValueKindSelection,
			selection: v.selection.Eq(i),
		}, nil
	default:
		panic("unexpected value kind")
	}
}

func (v *Value) Iterator() (Value, error) {
	nextVal := Value{
		kind: ValueKindIterator,
	}

	if v.kind == ValueKindIterator {
		for _, item := range v.array {
			if item.kind == ValueKindIterator {
				panic("only top value can be an iterator")
			}

			children, err := item.children()
			if err != nil {
				return Value{}, err
			}

			nextVal.iterator = append(nextVal.iterator, children...)
		}

		return nextVal, nil
	}

	var err error
	nextVal.iterator, err = v.children()
	if err != nil {
		return Value{}, err
	}

	return nextVal, nil
}

func (v *Value) children() ([]Value, error) {
	switch v.kind {
	case ValueKindArray:
		return v.array, nil
	case ValueKindSelection:
		ret := []Value{}
		v.selection.Each(func(_ int, s *goquery.Selection) {
			ret = append(ret, Value{
				kind:      ValueKindSelection,
				selection: s,
			})
		})
		return ret, nil
	case ValueKindString:
		return nil, errors.New("tried to iterate string")
	case ValueKindEmpty:
		return nil, errors.New("tried to iterate empty value")
	default:
		panic("children: unexpected value kind")
	}
}

func (v *Value) String() string {
	switch v.kind {
	case ValueKindArray:
		return string(must(json.Marshal(v)))
	case ValueKindIterator:
		buf := bytes.NewBuffer(nil)
		for i, item := range v.iterator {
			buf.WriteString(item.String())
			if i < len(v.iterator)-1 {
				buf.WriteString("\n")
			}
		}
		return buf.String()
	case ValueKindSelection:
		return must(v.selection.Html())
	case ValueKindString:
		return v.str
	case ValueKindEmpty:
		return ""
	default:
		panic("string: unexpected value kind")
	}
}

func (v *Value) MarshalJSON() ([]byte, error) {
	switch v.kind {
	case ValueKindArray:
		return json.Marshal(v.array)
	case ValueKindIterator, ValueKindSelection, ValueKindString:
		return json.Marshal(v.String())
	case ValueKindEmpty:
		return []byte("null"), nil
	default:
		panic("string: unexpected value kind")
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
