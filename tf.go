package tf

import (
	"fmt"
	"math"
	"unicode"
)

type Format uint8

const (
	String Format = iota
	Integer
	UnsignedInteger
	Float
	end
)

const (
	maxIterations = 100000
)

type position struct {
	row, col uint
	space    bool
	newline  bool
}

type TextField struct {
	cursor int        // cursor position in render slice
	render []position // text in screen system coordinate

	Text   []rune
	Format Format
}

func (t *TextField) cursorInRect() {
	if t.cursor < 0 {
		t.cursor = 0
	}
	if len(t.render) <= t.cursor {
		t.cursor = len(t.render) - 1
	}
}

func (t *TextField) CursorPosition() {
	fmt.Printf("HOLD")
}

func (t *TextField) CursorMoveUp() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if len(t.render) == 0 {
		return
	}
	if t.cursor == 0 {
		return
	}

	fmt.Printf("HOLD")
}

func (t *TextField) CursorMoveDown() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if len(t.render) == 0 {
		return
	}
	if t.cursor == len(t.render)-1 {
		return
	}
	for c := t.cursor + 1; c < len(t.render); c++ {
		if t.render[c].newline {
			continue
		}
		if t.render[t.cursor].row+1 == t.render[c].row &&
			t.render[t.cursor].col == t.render[c].col {
			t.cursor = c
			return
		}
		if t.render[t.cursor].row+2 == t.render[c].row {
			t.cursor = c - 1
			return
		}
	}
}

func (t *TextField) CursorMoveLeft() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if t.cursor == 0 {
		return
	}
	if len(t.render) == 0 {
		return
	}
	for 0 <= t.cursor-1 {
		t.cursor--
		if !t.render[t.cursor].newline {
			break
		}
	}
}

func (t *TextField) CursorMoveRight() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if t.cursor == len(t.render)-1 {
		return
	}
	if len(t.render) == 0 {
		return
	}
	for t.cursor+1 <= len(t.render)-1 {
		t.cursor++
		if !t.render[t.cursor].newline {
			break
		}
	}
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
func (t *TextField) SelectAll() { // DoubleClick
	fmt.Printf("HOLD")
}
func (t *TextField) InsertRune() { // runes and Enter
	fmt.Printf("HOLD")
}
func (t *TextField) RemoveBackspace() {
	fmt.Printf("HOLD")
}
func (t *TextField) RemoveDel() {
	fmt.Printf("HOLD")
}

func (t *TextField) Render(
	drawer func(row, col uint, r rune),
	cursor func(row, col uint),
) {
	for p := range t.render {
		if t.render[p].newline {
			continue
		}
		if t.render[p].space {
			drawer(t.render[p].row, t.render[p].col, '•')
			continue
		}
		drawer(t.render[p].row, t.render[p].col, t.Text[p])
	}
	if cursor != nil {
		if len(t.render) == 0 {
			cursor(0, 0)
		} else {
			cursor(t.render[t.cursor].row, t.render[t.cursor].col)
		}
	}
}

// runewidth is ignored.
//
// runes '\t', '\v', '\f', '\r', U+0085 (NEL), U+00A0 (NBSP) are iterpreted as '\n'.
//
func (t *TextField) SetWidth(width uint) {
	if width == 0 {
		t.render = nil // reset render
		return
	}
	// prepare render
	t.render = make([]position, len(t.Text))
	for i := range t.render {
		t.render[i].row = math.MaxUint
		t.render[i].col = math.MaxUint
	}

	pos := 0
	var row uint = 0
	for iter := 0; ; iter++ {
		if len(t.Text) <= pos {
			break
		}
		var col uint = 0
		for ; pos < len(t.Text); pos++ {
			// render
			t.render[pos] = position{row: row, col: col}
			//
			if t.Text[pos] == '\n' {
				t.render[pos].newline = true
				pos++
				break
			}
			if unicode.IsSpace(t.Text[pos]) && t.Text[pos] != ' ' {
				t.render[pos].space = true
				// pos++
				// break
			}
			if col == width {
				break
			}
			col++
		}
		row++
		if maxIterations < iter {
			panic(fmt.Errorf("iterations: %d. `%s` %#v", iter, string(t.Text), t))
		}
	}
}
