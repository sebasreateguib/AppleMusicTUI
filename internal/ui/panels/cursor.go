package panels

// cursor is the selection and scroll position of a list of single-line rows.
type cursor struct {
	index, top  int
	total, rows int
}

func (c *cursor) setTotal(total int) {
	c.total = total
	c.move(0)
}

func (c *cursor) setRows(rows int) {
	c.rows = max(1, rows)
	c.move(0)
}

func (c *cursor) set(index int) {
	c.index = index
	c.move(0)
}

// move changes the selection by delta and scrolls just enough to show it.
func (c *cursor) move(delta int) {
	c.index = max(0, min(c.index+delta, c.total-1))
	c.top = max(0, min(c.top, c.total-c.rows))
	c.top = min(c.top, c.index)
	c.top = max(c.top, c.index-c.rows+1)
}

// center scrolls so that the selection is in the middle.
func (c *cursor) center() {
	c.top = c.index - c.rows/2
	c.move(0)
}

// handleKey reports whether key moved the selection.
func (c *cursor) handleKey(key string) bool {
	before := c.index
	switch key {
	case "up", "k":
		c.move(-1)
	case "down", "j":
		c.move(1)
	case "pgup", "ctrl+u":
		c.move(-c.rows)
	case "pgdown", "ctrl+d":
		c.move(c.rows)
	case "home", "g":
		c.move(-c.total)
	case "end", "G":
		c.move(c.total)
	}
	return c.index != before
}

// window returns the visible rows [from, to).
func (c *cursor) window() (from, to int) {
	return c.top, min(c.total, c.top+c.rows)
}
