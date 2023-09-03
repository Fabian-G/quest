package todotxt

type Hook interface {
	Handle(prev *Item, current *Item)
}
