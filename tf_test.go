package tf

import (
	"bytes"
	"fmt"
	"testing"
)

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
		"\r",
		"\n",
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

const testdata = "testdata"

func Test(t *testing.T) {
	for ti := range txts {
		for wi := range widths {
			name := fmt.Sprintf("%04d-%04d-%04d", len(txts[ti]), ti, widths[wi])
			t.Run(name, func(t *testing.T) {
				check(t, string(txts[ti]), wi, name)
			})
		}
	}
}

func check(t *testing.T, str string, wi int, name string) {
	// prepare variables
	var (
		buf bytes.Buffer
		ta  = TextField{Text: []rune(str)}
	)
	// compare
	// defer func() {
	// 	var (
	// 		actual   = buf.Bytes()
	// 		filename = filepath.Join(testdata, name)
	// 	)
	// 	// for update test screens run in console:
	// 	// UPDATE=true go test
	// 	if os.Getenv("UPDATE") == "true" {
	// 		if err := ioutil.WriteFile(filename, actual, 0644); err != nil {
	// 			t.Fatalf("Cannot write snapshot to file: %v", err)
	// 		}
	// 	}
	// 	// get expect result
	// 	expect, err := ioutil.ReadFile(filename)
	// 	if err != nil {
	// 		t.Fatalf("Cannot read snapshot file: %v", err)
	// 	}
	// 	// compare
	// 	if !bytes.Equal(actual, expect) {
	// 		f2 := filename + ".new"
	// 		if err := ioutil.WriteFile(f2, actual, 0644); err != nil {
	// 			t.Fatalf("Cannot write snapshot to file new: %v", err)
	// 		}
	// 		t.Errorf("Snapshots is not same:\nActual:\n%s\nExpect:\n%s\nmeld %s %s",
	// 			actual,
	// 			expect,
	// 			filename, f2,
	// 		)
	// 	}
	// }()
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
		{name: "CursorPosition:left-top", f: func() {
			ta.CursorPosition(0, 0)
			ta.SetWidth(widths[wi])
		}}, // 7
		{name: "CursorPosition:right-top", f: func() {
			ta.CursorPosition(0, 100)
			ta.SetWidth(widths[wi])
		}}, // 8
		{name: "CursorPosition:left-bottom", f: func() {
			ta.CursorPosition(100, 0)
			ta.SetWidth(widths[wi])
		}}, // 9
		{name: "CursorPosition:right-bottom", f: func() {
			ta.CursorPosition(100, 100)
			ta.SetWidth(widths[wi])
		}}, // 10
		{name: "CursorPosition:1,1", f: func() {
			ta.CursorPosition(1, 1)
			ta.SetWidth(widths[wi])
		}}, // 11
		{name: "CursorPosition:100,1", f: func() {
			ta.CursorPosition(100, 1)
			ta.SetWidth(widths[wi])
		}}, // 12
		{name: "CursorPosition:1,100", f: func() {
			ta.CursorPosition(1, 100)
			ta.SetWidth(widths[wi])
		}}, // 13
		// {name: "CursorMoveHome", f: ta.CursorMoveHome},
		// {name: "CursorMoveEnd", f: ta.CursorMoveEnd},
		// {name: "CursorPageDown", f: ta.CursorPageDown},
		// {name: "CursorPageUp", f: ta.CursorPageUp},
	}
	var ms []movement

	// cursor position
	ms = append(ms, moves[7], moves[8], moves[9], moves[10],
		moves[11], moves[12], moves[13])
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
		ms = append(ms, moves[p])
	}
	ms = append(ms, repeat(3, moves[3])...)
	// cursor position
	ms = append(ms, moves[7], moves[8], moves[9], moves[10],
		moves[11], moves[12], moves[13])
	// render
	for i := range ms {
		fmt.Fprintf(&buf, "Move to: %s\n", ms[i].name)
		ta.SetWidth(widths[wi])
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

func TestInsert(t *testing.T) {
	tcs := []struct {
		input  string
		filter func(r rune) bool
		expect string
	}{
		{
			input:  "1o*5,2ds0qw.epp",
			filter: UnsignedInteger,
			expect: "1520",
		},
		{
			input:  "wefv-sdl;1o*5,2ds0qw.epp",
			filter: UnsignedInteger,
			expect: "1520",
		},
		{
			input:  "wefv-sdl;1o*5,2ds0qw.epp",
			filter: Integer,
			expect: "-1520",
		},
		{
			input:  "wefv+sdl;1o*5,2ds0qw.epp",
			filter: Integer,
			expect: "+1520",
		},
		{
			input:  "wfv+sdl;1o*5,2ds0qw.csscs3dfd4sdpp",
			filter: Float,
			expect: "+1520.34",
		},
		{
			input:  "+1.cvb232cvbevcb-cv0wcvb3",
			filter: Float,
			expect: "+1.232e-03",
		},
	}
	var width uint = 20
	for i := range tcs {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			ta := TextField{Text: []rune(""), Filter: tcs[i].filter}
			ta.SetWidth(width)
			for _, r := range []rune(tcs[i].input) {
				ta.Insert(r)
				ta.SetWidth(width)
			}
			var b Buffer
			ta.Render(b.Drawer, nil)
			actual := string(b.m[0])
			expect := tcs[i].expect
			if actual != expect {
				t.Errorf("result is not same: `%s`, but not `%s`",
					actual, expect)
			}
		})
	}
}

func TestCursor(t *testing.T) {
	var width uint = 20
	text := "1234\n12\n1234"
	ta := TextField{}
	tcs := []struct {
		name   string
		move   []func()
		expect []string
	}{
		{
			name: "Down",
			move: []func(){
				func() { ta.CursorPosition(0, 100) },
				func() { ta.CursorMoveDown() },
				func() { ta.CursorMoveDown() },
			},
			expect: []string{
				"1234█\n12\n1234\n",
				"1234\n12█\n1234\n",
				"1234\n12\n12█4\n",
			},
		},
		{
			name: "Up",
			move: []func(){
				func() { ta.CursorPosition(100, 100) },
				func() { ta.CursorMoveUp() },
				func() { ta.CursorMoveUp() },
			},
			expect: []string{
				"1234\n12\n1234█\n",
				"1234\n12█\n1234\n",
				"12█4\n12\n1234\n",
			},
		},
		{
			name: "Left",
			move: []func(){
				func() { ta.CursorPosition(1, 1) },
				func() { ta.CursorMoveLeft() },
				func() { ta.CursorMoveLeft() },
			},
			expect: []string{
				"1234\n1█\n1234\n",
				"1234\n█2\n1234\n",
				"1234█\n12\n1234\n",
			},
		},
		{
			name: "Right",
			move: []func(){
				func() { ta.CursorPosition(1, 1) },
				func() { ta.CursorMoveRight() },
				func() { ta.CursorMoveRight() },
			},
			expect: []string{
				"1234\n1█\n1234\n",
				"1234\n12█\n1234\n",
				"1234\n12\n█234\n",
			},
		},
		{
			name: "Insert",
			move: []func(){
				func() { ta.CursorPosition(1, 100) },
				func() { ta.Insert('W') },
				func() { ta.CursorMoveRight() },
				func() { ta.Insert('W') },
				func() { ta.CursorPosition(100, 100) },
				func() { ta.CursorMoveRight() },
				func() { ta.Insert('W') },
			},
			expect: []string{
				"1234\n12█\n1234\n",
				"1234\n12W█\n1234\n",
				"1234\n12W\n█234\n",
				"1234\n12W\nW█234\n",
				"1234\n12W\nW1234█\n",
				"1234\n12W\nW1234█\n",
				"1234\n12W\nW1234W█\n",
			},
		},
	}
	for i := range tcs {
		t.Run(tcs[i].name, func(t *testing.T) {
			ta.Text = []rune(text)
			ta.SetWidth(width)
			if len(tcs[i].move) != len(tcs[i].expect) {
				t.Errorf("not valid input data")
			}
			for p := range tcs[i].move {
				tcs[i].move[p]()
				if !ta.NoUpdate {
					ta.SetWidth(width)
				}
				var b Buffer
				ta.Render(b.Drawer, b.Cursor)
				actual := b.Text()
				if actual != tcs[i].expect[p] {
					t.Errorf("Step %2d\nresult is not same:\nActual:\n%s\nExpect:\n%s",
						p, actual, tcs[i].expect[p])
				}
			}
		})
	}
}
