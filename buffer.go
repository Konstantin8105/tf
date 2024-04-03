package tf

import "fmt"

var (
	defaultCursor = rune('█')
	errorRune     = rune('#')
)

type Buffer [][]rune

func (b *Buffer) Drawer(row, col uint, r rune) {
	for i := len(*b); i <= int(row); i++ {
		*b = append(*b, make([]rune, 0))
	}
	for i := len((*b)[row]); i <= int(col); i++ {
		(*b)[row] = append((*b)[row], errorRune)
	}
	(*b)[row][col] = r
}

func (b *Buffer) Cursor(row, col uint) {
	b.Drawer(row, col, defaultCursor)
}

func (b Buffer) String() string {
	var str string
	var w int
	for r := range b {
		str += fmt.Sprintf("%09d|", r+1)
		for c := range b[r] {
			str += string(b[r][c])
		}
		if width := len(b[r]); w < width {
			w = width
		}
		str += fmt.Sprintf("| width:%09d\n", len(b[r]))
	}
	str += fmt.Sprintf("rows  = %3d\n", len(b))
	str += fmt.Sprintf("width = %3d\n", w)
	return str
}

func (b Buffer) Text() string {
	var str string
	for r := range b {
		for c := range b[r] {
			str += string(b[r][c])
		}
		str += "\n"
	}
	return str
}

func (b Buffer) ErrorRune() bool {
	for r := range b {
		for c := range b[r] {
			if b[r][c] == errorRune {
				return true
			}
		}
	}
	return false
}

func (b Buffer) HasCursor() bool {
	found := false
	for r := range b {
		for c := range b[r] {
			if b[r][c] == defaultCursor {
				found = true
			}
		}
	}
	return found
}
