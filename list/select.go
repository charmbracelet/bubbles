package list

import "fmt"

// ToggleSelectCursor toggles the selected status
// of the current Index if amount is 0
// returns err != nil when amount lands outside list and safely does nothing
// else if amount is not 0 toggles selected amount items
// excluding the item on which the cursor would land
func (m *Model) ToggleSelectCursor(amount int) error {
	if m.Len() == 0 {
		return OutOfBounds(fmt.Errorf("No Items"))
	}
	if amount == 0 {
		m.listItems[m.viewPos.Cursor].selected = !m.listItems[m.viewPos.Cursor].selected
	}

	direction := 1
	if amount < 0 {
		direction = -1
	}

	cur := m.viewPos.Cursor

	target, err := m.MoveCursor(amount)
	start, end := cur, target
	if direction < 0 {
		start, end = target+1, cur+1
	}
	// mark/start at first item
	if cur+amount < 0 {
		start = 0
	}
	// mark last item when trying to go beyond list
	if cur+amount >= m.Len() {
		end++
	}
	for c := start; c < end; c++ {
		m.listItems[c].selected = !m.listItems[c].selected
	}
	return err
}

// ToggleSelect swaps the selected state of the item at the given index
// or returns a error if index is OutOfBounds.
func (m *Model) ToggleSelect(index int) error {
	i, err := m.ValidIndex(index)
	if err != nil {
		return err
	}
	m.listItems[i].selected = !m.listItems[i].selected
	return nil
}

// MarkSelectCursor selects or unselects depending on 'mark'
// amount = 0 changes the current item but does not move the cursor
// if amount would be outside the list error is from type OutOfBounds
// else all items till but excluding the end cursor position gets (un-)marked
func (m *Model) MarkSelectCursor(amount int, mark bool) error {
	cur := m.viewPos.Cursor
	direction := 1
	if amount < 0 {
		direction = -1
	}
	target := cur + amount - direction

	target, err := m.ValidIndex(target)
	if m.Len() == 0 {
		return err
	}
	// correct amount in case target has changed
	amount = target - cur + direction

	if amount == 0 {
		m.listItems[cur].selected = mark
		return nil
	}
	for c := 0; c < amount*direction; c++ {
		m.listItems[cur+c].selected = mark
	}
	m.viewPos.Cursor = target
	_, errSec := m.MoveCursor(direction)
	if err == nil {
		err = errSec
	}
	return err
}

// MarkSelect sets the selected state of the item at the given index to true
// or returns a error if index is OutOfBounds.
func (m *Model) MarkSelect(index int, mark bool) error {
	i, err := m.ValidIndex(index)
	if err != nil {
		return err
	}
	m.listItems[i].selected = mark
	return nil
}

// MarkSelectAll marks all items of the list according to mark
// or returns OutOfBounds if list has no Items
func (m *Model) MarkSelectAll(mark bool) error {
	_, err := m.ValidIndex(0)
	if m.Len() == 0 {
		return err
	}
	for c := range m.listItems {
		m.listItems[c].selected = mark
	}
	return err
}

// ToggleAllSelected inverts the select state of ALL items
func (m *Model) ToggleAllSelected() error {
	_, err := m.ValidIndex(0)
	if m.Len() == 0 {
		return err
	}
	for i := range m.listItems {
		m.listItems[i].selected = !m.listItems[i].selected
	}
	return err
}

// IsSelected returns true if the given Item is selected
// false otherwise. If the requested index is outside the list
// error is not nil.
func (m *Model) IsSelected(index int) (bool, error) {
	index, err := m.ValidIndex(index)
	if m.Len() == 0 {
		return false, err
	}
	return m.listItems[index].selected, err
}

// GetAllSelected returns you a list of all items
// that are selected in current (displayed) order
func (m *Model) GetAllSelected() []fmt.Stringer {
	var selected []fmt.Stringer
	for _, item := range m.listItems {
		if item.selected {
			selected = append(selected, item.value)
		}
	}
	return selected
}
