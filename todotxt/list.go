package todotxt

type List struct {
	tasks []*Item
	Hooks []Hook
}

func ListOf(items ...*Item) *List {
	newList := &List{}
	for _, item := range items {
		newList.Add(item)
	}
	return newList
}

func (l *List) Tasks() []*Item {
	return l.tasks
}

func (l *List) Add(item *Item) {
	l.tasks = append(l.tasks, item)
}

func (l *List) Get(idx int) *Item {
	return l.tasks[idx]
}

func (l *List) Len() int {
	return len(l.tasks)
}
