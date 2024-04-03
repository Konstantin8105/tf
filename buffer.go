package tf

import "fmt"

var (
	defaultCursor = rune('█')
	errorRune     = rune('#')
)

type Buffer struct {
	m [][]rune
}

func (b *Buffer) Drawer(row, col uint, r rune) {
	for i := len(b.m); i <= int(row); i++ {
		b.m = append(b.m, make([]rune, 0))
	}
	for i := len(b.m[row]); i <= int(col); i++ {
		b.m[row] = append(b.m[row], errorRune)
	}
	b.m[row][col] = r
}

func (b *Buffer) Cursor(row, col uint) {
	b.Drawer(row, col, defaultCursor)
}

func (b Buffer) String() string {
	var str string
	var w int
	for r := range b.m {
		str += fmt.Sprintf("%09d|", r+1)
		for c := range b.m[r] {
			str += string(b.m[r][c])
		}
		if width := len(b.m[r]); w < width {
			w = width
		}
		str += fmt.Sprintf("| width:%09d\n", len(b.m[r]))
	}
	str += fmt.Sprintf("rows  = %3d\n", len(b.m))
	str += fmt.Sprintf("width = %3d\n", w)
	return str
}

func (b Buffer) Text() string {
	var str string
	for r := range b.m {
		for c := range b.m[r] {
			str += string(b.m[r][c])
		}
		str += "\n"
	}
	return str
}

func (b Buffer) ErrorRune() bool {
	for r := range b.m {
		for c := range b.m[r] {
			if b.m[r][c] == errorRune {
				return true
			}
		}
	}
	return false
}

func (b Buffer) HasCursor() bool {
	found := false
	for r := range b.m {
		for c := range b.m[r] {
			if b.m[r][c] == defaultCursor {
				found = true
			}
		}
	}
	return found
}
