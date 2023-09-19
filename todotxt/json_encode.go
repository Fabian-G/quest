package todotxt

import (
	"bufio"
	"encoding/json"
	"io"
	"time"
)

var DefaultJsonEncoder = JsonEncoder{}

type JsonEncoder struct {
}

type jsonItem struct {
	Done             bool       `json:"done,omitempty"`
	Priority         string     `json:"priority,omitempty"`
	Tags             Tags       `json:"tags,omitempty"`
	Contexts         []Context  `json:"contexts,omitempty"`
	Project          []Project  `json:"project,omitempty"`
	Creation         *time.Time `json:"creation,omitempty"`
	Completion       *time.Time `json:"completion,omitempty"`
	Description      string     `json:"description,omitempty"`
	CleanDescription string     `json:"clean_description,omitempty"`
}

func (f JsonEncoder) Encode(w io.Writer, tasks []*Item) error {
	out := bufio.NewWriter(w)
	jsonItems := make([]jsonItem, 0, len(tasks))
	for _, t := range tasks {
		projects, contexts, tags := t.Projects(), t.Contexts(), t.Tags()
		jsonItems = append(jsonItems, jsonItem{
			Done:             t.Done(),
			Priority:         t.prio.String(),
			Tags:             tags,
			Contexts:         contexts,
			Project:          projects,
			Creation:         t.CreationDate(),
			Completion:       t.CompletionDate(),
			Description:      t.Description(),
			CleanDescription: t.CleanDescription(t.Projects(), t.Contexts(), tags.Keys()),
		})
	}
	json.NewEncoder(out).Encode(jsonItems)
	return out.Flush()
}
