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

type position struct{ row, col uint }

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

func (t *TextField) CursorPosition() {}
func (t *TextField) CursorMoveUp()   {}
func (t *TextField) CursorMoveDown() {}

func (t *TextField) CursorMoveLeft() {
	// cursor correction
	t.cursorInRect()
	defer t.cursorInRect()
	// action
	if t.cursor == 0 {
		return
	}
	for 0 <= t.cursor {
		t.cursor--
		if t.render[t.cursor].row != math.MaxUint &&
			t.render[t.cursor].col != math.MaxUint {
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
	for t.cursor <= len(t.render)-1 {
		t.cursor++
		if t.render[t.cursor].row != math.MaxUint &&
			t.render[t.cursor].col != math.MaxUint {
			break
		}
	}
}

func (t *TextField) CursorMoveHome()  {}
func (t *TextField) CursorMoveEnd()   {}
func (t *TextField) CursorPageDown()  {}
func (t *TextField) CursorPageUp()    {}
func (t *TextField) SelectAll()       {} // DoubleClick
func (t *TextField) InsertRune()      {} // runes and Enter
func (t *TextField) RemoveBackspace() {}
func (t *TextField) RemoveDel()       {}

func (t *TextField) Render(
	drawer func(row, col uint, r rune),
	cursor func(row, col uint),
) {
	for p := range t.render {
		if t.render[p].row == math.MaxUint || t.render[p].col == math.MaxUint {
			continue
		}
		if cursor != nil {
			if p == t.cursor {
				cursor(t.render[p].row, t.render[p].col)
				continue
			}
		}
		drawer(t.render[p].row, t.render[p].col, t.Text[p])
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
			if unicode.IsSpace(t.Text[pos]) && t.Text[pos] != ' ' {
				pos++
				break
			}
			if col == width {
				break
			}
			// render
			t.render[pos] = position{row: row, col: col}
			col++
		}
		row++
		if maxIterations < iter {
			panic(fmt.Errorf("iterations: %d. `%s` %#v", iter, string(t.Text), t))
		}
	}
}
