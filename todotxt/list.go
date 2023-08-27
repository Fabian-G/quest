package todotxt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type List []*Item

type ReadError struct {
	BaseError  error
	Line       string
	LineNumber int
}

func (r ReadError) Error() string {
	return fmt.Sprintf("could not read line number %d: %v\n\tThe problematic line was: %s", r.LineNumber, r.BaseError, r.Line)
}

func (l List) Save(w io.Writer, formatter Formatter) error {
	out := bufio.NewWriter(w)
	for _, item := range l {
		_, err := out.WriteString(formatter.Format(item) + "\n")
		if err != nil {
			return fmt.Errorf("could not write item %v: %w", item, err)
		}
	}
	return out.Flush()
}

func Read(r io.Reader) (List, error) {
	list := List(make([]*Item, 0))
	var errs []error
	var lineNumber int
	in := bufio.NewScanner(r)
	for in.Scan() {
		lineNumber++
		text := in.Text()
		item, err := ParseItem(text)
		if err != nil {
			errs = append(errs, ReadError{
				BaseError:  err,
				Line:       text,
				LineNumber: lineNumber,
			})
		}
		list = append(list, item)
	}
	if err := in.Err(); err != nil {
		return nil, fmt.Errorf("could not read input: %w", err)
	}
	return list, errors.Join(errs...)
}
