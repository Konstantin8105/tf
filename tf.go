package tf

import (
	"fmt"
	"math"
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

func (t *TextField) CursorPosition() {}
func (t *TextField) CursorMoveUp()   {}
func (t *TextField) CursorMoveDown() {}
func (t *TextField) CursorMoveLeft() {
	t.cursor--
	if t.cursor < 0 {
		t.cursor = 0
		return
	}
	if t.render[t.cursor].row == math.MaxUint || t.render[t.cursor].col == math.MaxUint {
		t.CursorMoveLeft()
	}
}
func (t *TextField) CursorMoveRight() {
	t.cursor++
	if len(t.render) <= t.cursor {
		t.cursor--
		return
	}
	if t.render[t.cursor].row == math.MaxUint || t.render[t.cursor].col == math.MaxUint {
		t.CursorMoveRight()
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

// runewidth is ignored
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
			if (t.Text)[pos] == '\n' {
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
