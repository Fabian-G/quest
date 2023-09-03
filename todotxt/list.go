package todotxt

type List struct {
	tasks         []*Item
	hooksDisabled bool
	Hooks         []Hook
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
	item.emitFunc = l.emit
	l.tasks = append(l.tasks, item)
	l.emit(ModEvent{Previous: nil, Current: item})
}

func (l *List) Get(idx int) *Item {
	return l.tasks[idx]
}

func (l *List) Len() int {
	return len(l.tasks)
}

func (l *List) emit(me ModEvent) {
	if l.hooksDisabled {
		return
	}

	l.hooksDisabled = true
	defer func() {
		l.hooksDisabled = false
	}()
	for _, h := range l.Hooks {
		h.Handle(me)
	}
}
