package tf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var formats []Format

func init() {
	for i := 0; i < int(end); i++ {
		formats = append(formats, Format(i))
	}
}

var txts [][]rune

func init() {
	for _, str := range []string{
		"",
		"foo",
		"Lorem ipsum dolor \nsit amet.",
		"Lorem ipsum dolor \n\nsit amet.",
		"Sed maximus tempor condimentum.\n\nNam et risus est. Cras ornare iaculis orci, \n\nsit amet fringilla nisl pharetra quis.",
		`See: https://ru.wikipedia.org/wiki/Go
	Go (часто также Golang) — компилируемый многопоточный язык программирования, разработанный внутри компании Google[4]. Разработка Go началась в сентябре 2007 года, его непосредственным проектированием занимались Роберт Гризмер, Роб Пайк и Кен Томпсон[5], занимавшиеся до этого проектом разработки операционной системы Inferno. Официально язык был представлен в ноябре 2009 года.
	На данный момент поддержка официального компилятора, разрабатываемого создателями языка, осуществляется для операционных систем FreeBSD, OpenBSD, Linux, macOS, Windows, DragonFly BSD, Plan 9, Solaris, Android, AIX.[6].
`,
		`See: https://golang.org/
// You can edit this code!
// Click here and start typing.
package main
import "fmt"
func main() {
	// 012345689012345678901234567890
	fmt.Println("[red]Hello, 世界[white]")
}
You will see:
世界
`,
		"世界",
		"\t\n\r\t",
	} {
		txts = append(txts, []rune(str))
	}
}

var widths = []uint{0, 1, 2, 3, 4, 10, 25, 50, 100}

const errorRune = rune('#')

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

const defaultCursor = '█'

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

const testdata = "testdata"

func Test(t *testing.T) {
	for ti := range txts {
		for wi := range widths {
			name := fmt.Sprintf("%04d-%04d", len(txts[ti]), widths[wi])
			t.Run(name, func(t *testing.T) {
				check(t, ti, wi, name)
			})
		}
	}
}

func check(t *testing.T, ti, wi int, name string) {
	// prepare variables
	var (
		buf bytes.Buffer
		f   = String
		ta  = TextField{Text: txts[ti], Format: f}
	)
	// compare
	defer func() {
		var (
			actual   = buf.Bytes()
			filename = filepath.Join(testdata, name)
		)
		// for update test screens run in console:
		// UPDATE=true go test
		if os.Getenv("UPDATE") == "true" {
			if err := ioutil.WriteFile(filename, actual, 0644); err != nil {
				t.Fatalf("Cannot write snapshot to file: %v", err)
			}
		}
		// get expect result
		expect, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("Cannot read snapshot file: %v", err)
		}
		// compare
		if !bytes.Equal(actual, expect) {
			f2 := filename + ".new"
			if err := ioutil.WriteFile(f2, actual, 0644); err != nil {
				t.Fatalf("Cannot write snapshot to file new: %v", err)
			}
			t.Errorf("Snapshots is not same:\nActual:\n%s\nExpect:\n%s\nmeld %s %s",
				actual,
				expect,
				filename, f2,
			)
		}
	}()
	// test: static
	var s1 string
	{
		fmt.Fprintf(&buf, "Test: Static\n")
		ta.SetWidth(widths[wi])
		var b Buffer
		ta.Render(b.Drawer, b.Cursor)
		if b.ErrorRune() {
			t.Fatalf("buffer have error rune")
		}
		if !b.HasCursor() {
			t.Fatalf("no cursor")
		}
		fmt.Fprintf(&buf, "%s\n", b)
		s1 = b.String()
	}
	// test: resize to less
	if wi == 0 {
		return
	}
	{
		fmt.Fprintf(&buf, "Test: resize to less\n")
		ta.SetWidth(widths[wi-1])
		var b Buffer
		ta.Render(b.Drawer, b.Cursor)
		if b.ErrorRune() {
			t.Fatalf("buffer have error rune")
		}
		if !b.HasCursor() {
			t.Fatalf("no cursor")
		}
		fmt.Fprintf(&buf, "%s\n", b)
	}
	// test: resize to more
	var s2 string
	{
		fmt.Fprintf(&buf, "Test: resize to more\n")
		ta.SetWidth(widths[wi])
		var b Buffer
		ta.Render(b.Drawer, b.Cursor)
		if b.ErrorRune() {
			t.Fatalf("buffer have error rune")
		}
		if !b.HasCursor() {
			t.Fatalf("no cursor")
		}
		fmt.Fprintf(&buf, "%s\n", b)
		s2 = b.String()
	}
	// compare resizes
	if s1 != s2 {
		t.Errorf("resize not valid:\n%s\n%s", s1, s2)
	}
	// cursor move
	type movement struct {
		name string
		f    func()
	}
	repeat := func(number int, m movement) (ms []movement) {
		ms = make([]movement, number)
		for i := 0; i < number; i++ {
			ms[i].name = m.name
			ms[i].f = m.f
		}
		return
	}
	moves := []movement{
		{name: "CursorMoveUp", f: ta.CursorMoveUp},       // 0
		{name: "CursorMoveDown", f: ta.CursorMoveDown},   // 1
		{name: "CursorMoveLeft", f: ta.CursorMoveLeft},   // 2
		{name: "CursorMoveRight", f: ta.CursorMoveRight}, // 3
		{name: "InsertRuneA", f: func() {
			ta.Insert('W')
			ta.SetWidth(widths[wi])
		}}, // 4
		{name: "KeyBackspace", f: func() {
			ta.KeyBackspace()
			ta.SetWidth(widths[wi])
		}}, // 5
		{name: "KeyDel", f: func() {
			ta.KeyDel()
			ta.SetWidth(widths[wi])
		}}, // 6
		// {name: "CursorMoveHome", f: ta.CursorMoveHome},
		// {name: "CursorMoveEnd", f: ta.CursorMoveEnd},
		// {name: "CursorPageDown", f: ta.CursorPageDown},
		// {name: "CursorPageUp", f: ta.CursorPageUp},
	}
	var ms []movement

	// right - left, down - up
	ms = append(ms, repeat(4, moves[3])...)
	ms = append(ms, repeat(5, moves[2])...)
	ms = append(ms, repeat(4, moves[1])...)
	ms = append(ms, repeat(5, moves[0])...)
	// square
	ms = append(ms, repeat(4, moves[3])...)
	ms = append(ms, repeat(4, moves[1])...)
	ms = append(ms, repeat(4, moves[2])...)
	ms = append(ms, repeat(4, moves[0])...)
	// on bottom and right corner
	ms = append(ms, repeat(10, moves[1])...)
	ms = append(ms, repeat(10, moves[3])...)
	// insert rune
	ms = append(ms,
		moves[3], moves[4],
		moves[2], moves[4],
		moves[1], moves[4],
		moves[0], moves[4],
	)
	// backspace
	ms = append(ms,
		moves[3], moves[5],
		moves[2], moves[5],
		moves[1], moves[5],
		moves[0], moves[5],
	)
	// del
	ms = append(ms,
		moves[3], moves[6],
		moves[2], moves[6],
		moves[1], moves[6],
		moves[0], moves[6],
	)
	// all moves
	ms = append(ms, moves[5], moves[6], moves[5], moves[6])
	for p := range moves {
		ms = append(ms,  moves[p])
	}
	ms = append(ms, repeat(3, moves[3])...)
	for i := range ms {
		fmt.Fprintf(&buf, "Move to: %s\n", ms[i].name)
		ms[i].f()
		var b Buffer
		ta.Render(b.Drawer, b.Cursor)
		if b.ErrorRune() {
			t.Fatalf("buffer have error rune for move cursor")
		}
		if !b.HasCursor() {
			t.Fatalf("no cursor")
		}
		fmt.Fprintf(&buf, "%s\n", b)
	}
}
