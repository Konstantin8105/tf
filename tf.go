package tf

import (
	"fmt"
	"unicode"
)

type Format uint8

func UnsignedInteger(r rune) (insert bool) {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

func Integer(r rune) (insert bool) {
	if UnsignedInteger(r) {
		return true
	}
	switch r {
	case '+', '-':
		return true
	}
	return false
}

func Float(r rune) (insert bool) {
	if Integer(r) {
		return true
	}
	switch r {
	case '.', 'e', 'E':
		return true
	}
	return false
}

const (
	maxIterations = 100000
)

type symType uint8

const (
	symbol  symType = iota // 0
	space                  // 1
	newline                // 2
	endtext                // 3
)

type position struct {
	row, col uint
	t        symType
}

type TextField struct {
	cursor int        // cursor position in render slice
	render []position // text in screen system coordinate

	Text     []rune
	Filter   func(r rune) (insert bool)
	NoUpdate bool
}

func (t *TextField) cursorInRect() {
	if len(t.render) == 0 {
		panic(fmt.Errorf("not valid. Try run SetWidth: %#v %#v", t.render, t.Text))
	}
	if 0 < len(t.render) && len(t.render) <= int(t.cursor) {
		t.cursor = (len(t.render)) - 1
	}
}

func (t *TextField) CursorPosition(row, col uint) {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	// find cursor position
	if row == 0 && col == 0 {
		t.cursor = 0
		return
	}
	if last := len(t.render) - 1; t.render[last].row <= row &&
		t.render[last].col <= col {
		t.cursor = last
		return
	}
	for i := range t.render {
		if t.render[i].row == row && t.render[i].col == col {
			t.cursor = i
			return
		}
		if t.render[i].row == row+1 {
			t.cursor = i - 1
			return
		}
	}
	for i := len(t.render) - 1; 0 <= i; i-- {
		if t.render[i].col == col {
			t.cursor = i
			return
		}
	}
	panic(fmt.Errorf(
		"cursor is not found: %v %v %v",
		row, col, t.render[len(t.render)-1],
	))
}

func (t *TextField) CursorMoveUp() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if t.cursor == 0 {
		return
	}
	for c := t.cursor - 1; 0 <= c; c-- {
		if t.render[t.cursor].row-1 == t.render[c].row &&
			t.render[c].col <= t.render[t.cursor].col {
			t.cursor = c
			return
		}
	}
}

func (t *TextField) CursorMoveDown() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if t.cursor == len(t.render)-1 {
		return
	}
	t.CursorPosition(t.render[t.cursor].row+1, t.render[t.cursor].col)
}

func (t *TextField) CursorMoveLeft() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if t.cursor == 0 {
		return
	}
	t.cursor--
}

func (t *TextField) CursorMoveRight() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if t.cursor == len(t.render)-1 {
		return
	}
	t.cursor++
}

// func (t *TextField) CursorMoveHome() {
// 	fmt.Printf("HOLD")
// }
// func (t *TextField) CursorMoveEnd() {
// 	fmt.Printf("HOLD")
// }
// func (t *TextField) CursorPageDown() {
// 	fmt.Printf("HOLD")
// }
// func (t *TextField) CursorPageUp() {
// 	fmt.Printf("HOLD")
// }
// func (t *TextField) SelectAll() { // DoubleClick
// 	fmt.Printf("HOLD")
// }

// Insert rune, key Enter `\n` in text without update buffer.
// After that function run `SetWidth` for update buffer.
func (t *TextField) Insert(r rune) {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if t.Filter != nil && !t.Filter(r) {
		return
	}
	// Is need update?
	defer func() {
		t.cursor++
		t.NoUpdate = false
	}()
	if t.cursor == 0 {
		t.Text = append([]rune{r}, t.Text...)
		t.render = append([]position{{row: 0, col: 0, t: symbol}}, t.render...)
		return
	}
	symT := convert(r)
	// 	var addrow uint = 0
	// 	if symT == newline {
	// 		addrow = 1
	// 	}
	// 	if t.render[t.cursor].t == endtext {
	// 		fmt.Println("=====", addrow)
	// 		t.Text = append(t.Text, r)
	// 		var row, col uint
	// 		row = t.render[t.cursor].row
	// 		col = t.render[t.cursor].col
	//
	// 		if symT == newline {
	// 			row += 1
	// 			col = 0
	// 		}
	//
	// 		t.render[len(t.render)-1].t = symT
	// 		t.render[len(t.render)-1].row = row
	// 		t.render = append(t.render, position{row: row, col: col+1, t: endtext})
	// 		return
	// 	}
	t.Text = append(t.Text[:t.cursor], append([]rune{r}, t.Text[t.cursor:]...)...)
	t.render = append(t.render[:t.cursor], append([]position{
		position{row: 0, col: 0, t: symT},
	}, t.render[t.cursor:]...)...)
}

func convert(r rune) symType {
	if r == '\n' {
		return newline
	} else if unicode.IsSpace(r) && r != ' ' {
		return space
	}
	return symbol
}

func (t *TextField) KeyBackspace() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if len(t.render) == 0 {
		return
	}
	if t.cursor < 1 {
		return
	}
	// Is need update?
	defer func() {
		t.NoUpdate = false
	}()
	t.Text = append(t.Text[:t.cursor-1], t.Text[t.cursor:]...)
	t.cursor--
}

func (t *TextField) KeyDel() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if len(t.render) == 0 {
		return
	}
	// Is need update?
	defer func() {
		t.NoUpdate = false
	}()
	if len(t.render) == t.cursor+1 {
		// nothing to do
		return
	}
	t.Text = append(t.Text[:t.cursor], t.Text[t.cursor+1:]...)
}

func (t *TextField) Render(
	drawer func(row, col uint, r rune),
	cursor func(row, col uint),
) (height uint) {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	for p := range t.render {
		switch t.render[p].t {
		case symbol:
			drawer(t.render[p].row, t.render[p].col, t.Text[p])
		case space:
			drawer(t.render[p].row, t.render[p].col, '•')
		case newline:
			// drawer(t.render[p].row, t.render[p].col, '↵')
		case endtext:
			// drawer(t.render[p].row, t.render[p].col, 'X')
		default:
			panic(fmt.Errorf("undefined render symbol: %d", t.render[p].t))
		}
	}
	if cursor != nil {
		cursor(t.render[t.cursor].row, t.render[t.cursor].col)
	}

	return t.render[len(t.render)-1].row + 1
}

// runewidth is ignored.
//
// runes '\t', '\v', '\f', '\r', U+0085 (NEL), U+00A0 (NBSP) are iterpreted as '\n'.
//
func (t *TextField) SetWidth(width uint) {
	// Minimal width of text is:
	// 1 symbol - rune
	// 2 symbol - cursor
	const minWidth = 2
	if width < minWidth {
		t.render = []position{{row: 0, col: 0, t: endtext}} // reset render
		return
	}
	// change width for cursor place
	width -= 1
	// update text
	if t.NoUpdate {
		return
	}
	defer func() {
		t.NoUpdate = false
	}()
	// allocation
	{
		// last is endtext
		size := len(t.Text) + 1
		if size < len(t.render) {
			t.render = t.render[:size]
		}
		if size != len(t.render) {
			t.render = make([]position, size)
		}
	}

	// prepare render types
	for i := range t.Text {
		t.render[i].t = convert(t.Text[i])
	}
	// render rows, cols calculations
	var row, col uint
	for i := range t.Text {
		t.render[i].row = row
		t.render[i].col = col
		col++
		if t.render[i].t == newline || col == width {
			row++
			col = 0
		}
	}
	t.render[len(t.render)-1] = position{row: row, col: col, t: endtext}
}

func (t *TextField) GetRenderHeight() uint {
	last := len(t.render) - 1
	if last < 0 {
		return 0
	}
	return t.render[last].row + 1
}

func (t *TextField) GetRenderWidth() uint {
	w := uint(1)
	for i := range t.render {
		if w < t.render[i].col {
			w = t.render[i].col
		}
	}
	return w
}
