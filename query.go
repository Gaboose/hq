package main

type Filter struct {
	Comma []Comma `@@ ("," @@)*`
}

type Comma struct {
	Pipe []Pipe `@@ ("|" @@)*`
}

type Pipe struct {
	Array *Filter `("[" @@ "]"`
	Find  string  `| ("find" "(" @String ")")`
	Attr  string  `| ("attr" "(" @String ")")`
	Html  bool    `| @"html"`
	Text  bool    `| @"text")`

	Suffix Suffix `@@?`
}

type Suffix struct {
	Iterator bool `@("[" "]")`
	Index    *int `| "[" @("-"? Int) "]"`
}

func (p *Filter) Exec(val Value) (Value, error) {
	nextVals := []Value{}

	for _, c := range p.Comma {
		nextVal, err := c.Exec(val)
		if err != nil {
			return Value{}, err
		}

		nextVals = append(nextVals, nextVal)
	}

	if len(nextVals) == 1 {
		return nextVals[0], nil
	}

	return Value{
		kind:     ValueKindIterator,
		iterator: nextVals,
	}, nil
}

func (f *Comma) Exec(val Value) (Value, error) {
	var err error

	for _, c := range f.Pipe {
		var nextVal Value

		if val.kind == ValueKindIterator {
			nextVal.kind = ValueKindIterator

			for _, item := range val.iterator {
				nextItem, err := c.Exec(item)
				if err != nil {
					return Value{}, err
				}

				nextVal.iterator = append(nextVal.iterator, nextItem)
			}
		} else {
			nextVal, err = c.Exec(val)
			if err != nil {
				return Value{}, err
			}
		}

		val = nextVal
	}

	return val, nil
}

func (f *Pipe) Exec(val Value) (Value, error) {
	var err error

	switch {
	case f.Array != nil:
		val, err = f.Array.Exec(val)
		if err != nil {
			return Value{}, err
		}

		val = val.Array()
	case f.Find != "":
		val = val.Find(f.Find)
	case f.Attr != "":
		val = val.Attr(f.Attr)
	case f.Html:
		val, err = val.Html()
	case f.Text:
		val = val.Text()
	}

	if err != nil {
		return Value{}, err
	}

	switch {
	case f.Suffix.Iterator:
		val, err = val.Iterator()
	case f.Suffix.Index != nil:
		val, err = val.Index(*f.Suffix.Index)
	}

	return val, err
}
