package tabs

func NewTab(title string) Item {
	t := Item{}
	t.Title = title
	t.Active = false
	return t
}

type Item struct {
	Title  string
	Active bool
}
