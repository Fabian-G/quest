package todotxt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type List []*Item

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
	in := bufio.NewScanner(r)
	for in.Scan() {
		text := in.Text()
		item, err := ParseItem(text)
		errs = append(errs, err)
		list = append(list, item)
	}
	if err := in.Err(); err != nil {
		return nil, fmt.Errorf("could not read input: %w", err)
	}
	return list, errors.Join(errs...)
}
