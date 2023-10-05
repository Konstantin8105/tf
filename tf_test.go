package tf

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
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
		ta  = TextField{text: []rune(str)}
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
			ta := TextField{text: []rune(""), Filter: tcs[i].filter}
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
	tcs := []struct {
		name   string
		text   string
		move   []func(fake)
		expect []string
		eWidth []int
	}{
		{
			name: "Down",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 100) },
				func(ta fake) { ta.CursorMoveDown() },
				func(ta fake) { ta.CursorMoveDown() },
			},
			expect: []string{
				"1234█\n12\n1234\n",
				"1234\n12█\n1234\n",
				"1234\n12\n12█4\n",
			},
			eWidth: []int{4, 4, 4},
		},
		{
			name: "Up",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(100, 100) },
				func(ta fake) { ta.CursorMoveUp() },
				func(ta fake) { ta.CursorMoveUp() },
			},
			expect: []string{
				"1234\n12\n1234█\n",
				"1234\n12█\n1234\n",
				"12█4\n12\n1234\n",
			},
			eWidth: []int{4, 4, 4},
		},
		{
			name: "Left",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(1, 1) },
				func(ta fake) { ta.CursorMoveLeft() },
				func(ta fake) { ta.CursorMoveLeft() },
			},
			expect: []string{
				"1234\n1█\n1234\n",
				"1234\n█2\n1234\n",
				"1234█\n12\n1234\n",
			},
			eWidth: []int{4, 4, 4},
		},
		{
			name: "Right",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(1, 1) },
				func(ta fake) { ta.CursorMoveRight() },
				func(ta fake) { ta.CursorMoveRight() },
			},
			expect: []string{
				"1234\n1█\n1234\n",
				"1234\n12█\n1234\n",
				"1234\n12\n█234\n",
			},
			eWidth: []int{4, 4, 4},
		},
		{
			name: "Insert",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(1, 100) },
				func(ta fake) { ta.Insert('W') },
				func(ta fake) { ta.CursorMoveRight() },
				func(ta fake) { ta.Insert('W') },
				func(ta fake) { ta.CursorPosition(100, 100) },
				func(ta fake) { ta.CursorMoveRight() },
				func(ta fake) { ta.Insert('W') },
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
			eWidth: []int{4, 4, 4, 5, 5, 5, 6},
		},
		{
			name: "Down2",
			text: "123456\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 100) },
				func(ta fake) { ta.CursorMoveDown() },
			},
			expect: []string{
				"123456█\n1234\n",
				"123456\n1234█\n",
			},
			eWidth: []int{6, 6},
		},
		{
			name: "Enter",
			text: "123456",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 100) },
				func(ta fake) { ta.Insert('\n') },
				func(ta fake) { ta.Insert('\n') },
			},
			expect: []string{
				"123456█\n",
				"123456\n█\n",
				"123456\n\n█\n",
			},
			eWidth: []int{6, 6, 6},
		},
		{
			name: "Backspace",
			text: "123456",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 100) },
				func(ta fake) { ta.KeyBackspace() },
				func(ta fake) { ta.KeyBackspace() },
			},
			expect: []string{
				"123456█\n",
				"12345█\n",
				"1234█\n",
			},
			eWidth: []int{6, 5, 4},
		},
		{
			name: "Backspace2",
			text: "123456",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 3) },
				func(ta fake) { ta.KeyBackspace() },
				func(ta fake) { ta.KeyBackspace() },
			},
			expect: []string{
				"123█56\n",
				"12█56\n",
				"1█56\n",
			},
			eWidth: []int{6, 5, 4},
		},
		{
			name: "Del",
			text: "123456",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 0) },
				func(ta fake) { ta.KeyDel() },
				func(ta fake) { ta.KeyDel() },
			},
			expect: []string{
				"█23456\n",
				"█3456\n",
				"█456\n",
			},
			eWidth: []int{6, 5, 4},
		},
	}
	for i := range tcs {
		t.Run(tcs[i].name, func(t *testing.T) {
			ta := TextField{}
			text := []rune(tcs[i].text)
			ta.SetText(text)
			if string(text) != string(ta.GetText()){
				t.Errorf("not valid GetText")
			}
			ta.SetWidth(width)
			if len(tcs[i].move) != len(tcs[i].expect) {
				t.Errorf("not valid input data")
			}
			for p := range tcs[i].move {
				tcs[i].move[p](&ta)
				var b Buffer
				ta.Render(b.Drawer, b.Cursor)
				actual := b.Text()
				if actual != tcs[i].expect[p] {
					t.Errorf("Step %2d\nresult is not same:\nActual:\n%s\nExpect:\n%s",
						p, actual, tcs[i].expect[p])
				}
				// check height of render
				h := int(ta.GetRenderHeight())
				eh := strings.Count(tcs[i].expect[p], "\n")
				if h != eh {
					t.Errorf("not valid height of render: %d != %d", h, eh)
				}
				// check width of render
				actualSize := int(ta.GetRenderWidth())
				if actualSize != tcs[i].eWidth[p] {
					t.Errorf("not valid width of render: %d != %d",
						actualSize, tcs[i].eWidth[p])
				}
			}
		})
	}
}

func TestSet(t *testing.T) {
	var largetext string = "some text"
	for i := 0; i < 8; i++ {
		largetext += largetext
	}

	var wg sync.WaitGroup
	wg.Add(3)
	ta := TextField{}
	go func() {
		for i := range largetext {
			ta.SetText([]rune(largetext[:i]))
		}
		for i := len(largetext) - 1; 0 <= i; i-- {
			ta.SetText([]rune(largetext[:i]))
		}
		wg.Done()
	}()
	go func() {
		size := len(largetext) * 10
		for i := 0; i < size; i++ {
			ta.SetWidth(6)
		}
		wg.Done()
	}()
	go func() {
		size := len(largetext) * 10
		for i := 0; i < size; i++ {
			ta.SetWidth(10)
		}
		wg.Done()
	}()
	t.Logf("lenght: %d", len(largetext))
	wg.Wait()
}

type fake interface {
	CursorPosition(row, col uint)
	CursorMoveUp()
	CursorMoveDown()
	CursorMoveLeft()
	CursorMoveRight()

	Insert(r rune)
	KeyBackspace()
	KeyDel()
}

func TestSingleLine(t *testing.T) {
	var width uint = 8
	tcs := []struct {
		name   string
		text   string
		move   []func(fake)
		expect []string
		eWidth []int
	}{
		{
			name: "Down",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 100) },
				func(ta fake) { ta.CursorMoveDown() },
				func(ta fake) { ta.CursorMoveDown() },
			},
			expect: []string{
				"1234█\n", // 0
				"12█\n",   // 1
				"12█4\n",  // 2
			},
			eWidth: []int{4, 4, 4},
		},
		{
			name: "Up",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(100, 100) },
				func(ta fake) { ta.CursorMoveUp() },
				func(ta fake) { ta.CursorMoveUp() },
			},
			expect: []string{
				"1234█\n",
				"12█\n",
				"12█4\n",
			},
			eWidth: []int{4, 4, 4},
		},
		{
			name: "Left",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(1, 1) },
				func(ta fake) { ta.CursorMoveLeft() },
				func(ta fake) { ta.CursorMoveLeft() },
			},
			expect: []string{
				"1█\n",
				"█2\n",
				"1234█\n",
			},
			eWidth: []int{4, 4, 4},
		},
		{
			name: "Right",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(1, 1) },
				func(ta fake) { ta.CursorMoveRight() },
				func(ta fake) { ta.CursorMoveRight() },
			},
			expect: []string{
				"1█\n",
				"12█\n",
				"█234\n",
			},
			eWidth: []int{4, 4, 4},
		},
		{
			name: "Insert",
			text: "1234\n12\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(1, 100) },
				func(ta fake) { ta.Insert('W') },
				func(ta fake) { ta.CursorMoveRight() },
				func(ta fake) { ta.Insert('W') },
				func(ta fake) { ta.CursorPosition(100, 100) },
				func(ta fake) { ta.CursorMoveRight() },
				func(ta fake) { ta.Insert('W') },
			},
			expect: []string{
				"12█\n",
				"12W█\n",
				"█234\n",
				"W█234\n",
				"W1234█\n",
				"W1234█\n",
				"W1234W█\n",
			},
			eWidth: []int{4, 4, 4, 5, 5, 5, 6},
		},
		{
			name: "Down2",
			text: "123456\n1234",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 100) },
				func(ta fake) { ta.CursorMoveDown() },
			},
			expect: []string{
				"123456█\n",
				"1234█\n",
			},
			eWidth: []int{6, 6},
		},
		{
			name: "Enter",
			text: "123456",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 100) },
				func(ta fake) { ta.Insert('\n') },
				func(ta fake) { ta.Insert('\n') },
			},
			expect: []string{
				"123456█\n",
				"█\n",
				"█\n",
			},
			eWidth: []int{6, 6, 6},
		},
		{
			name: "Backspace",
			text: "123456",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 100) },
				func(ta fake) { ta.KeyBackspace() },
				func(ta fake) { ta.KeyBackspace() },
			},
			expect: []string{
				"123456█\n",
				"12345█\n",
				"1234█\n",
			},
			eWidth: []int{6, 5, 4},
		},
		{
			name: "Backspace2",
			text: "123456",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 3) },
				func(ta fake) { ta.KeyBackspace() },
				func(ta fake) { ta.KeyBackspace() },
			},
			expect: []string{
				"123█56\n",
				"12█56\n",
				"1█56\n",
			},
			eWidth: []int{6, 5, 4},
		},
		{
			name: "Del",
			text: "123456",
			move: []func(fake){
				func(ta fake) { ta.CursorPosition(0, 0) },
				func(ta fake) { ta.KeyDel() },
				func(ta fake) { ta.KeyDel() },
			},
			expect: []string{
				"█23456\n",
				"█3456\n",
				"█456\n",
			},
			eWidth: []int{6, 5, 4},
		},
	}
	for i := range tcs {
		t.Run(tcs[i].name, func(t *testing.T) {
			ta := TextFieldLimit{}
			ta.SetLinesLimit(1)
			ta.SetText([]rune(tcs[i].text))
			ta.SetWidth(width)
			if len(tcs[i].move) != len(tcs[i].expect) {
				t.Errorf("not valid input data")
			}
			for p := range tcs[i].move {
				t.Logf("Step %2d", p)
				tcs[i].move[p](&ta)
				var b Buffer
				ta.Render(b.Drawer, b.Cursor)
				actual := b.Text()
				if actual != tcs[i].expect[p] {
					t.Errorf("result is not same:\nActual:\n%s\nExpect:\n%s",
						actual, tcs[i].expect[p])
				} else {
					t.Logf("%s", actual)
				}
				// check height of render
				h := int(ta.GetRenderHeight())
				eh := strings.Count(tcs[i].expect[p], "\n")
				if h != eh {
					t.Errorf("not valid height of render: %d != %d", h, eh)
				}
				// check width of render
				actualSize := int(ta.GetRenderWidth())
				if actualSize != tcs[i].eWidth[p] {
					t.Errorf("not valid width of render: %d != %d",
						actualSize, tcs[i].eWidth[p])
				}
			}
		})
	}
}

// goos: linux
// goarch: amd64
// pkg: github.com/Konstantin8105/tf
// cpu: Intel(R) Xeon(R) CPU E3-1240 V2 @ 3.40GHz
// Benchmark/Render-0637-0100-4         	  443305	      2528 ns/op	       0 B/op	       0 allocs/op
// Benchmark/Width-0637-0100-4          	  151008	      7783 ns/op	       0 B/op	       0 allocs/op
// Benchmark/RWNoChange-0637-0100-4     	  118446	     10170 ns/op	       0 B/op	       0 allocs/op
// Benchmark/RWChanged-0637-0100-4      	  116136	     10214 ns/op	       0 B/op	       0 allocs/op
//
// Benchmark/Render-0637-0100-4         	  477376	      2455 ns/op	       0 B/op	       0 allocs/op
// Benchmark/Width-0637-0100-4          	315199640	         3.960 ns/op	       0 B/op	       0 allocs/op
// Benchmark/RWNoChange-0637-0100-4     	  485542	      2515 ns/op	       0 B/op	       0 allocs/op
// Benchmark/RWChanged-0637-0100-4      	  108307	     10703 ns/op	       0 B/op	       0 allocs/op
// Benchmark/RSTNoChange-0637-0100-4    	  360322	      3101 ns/op	       0 B/op	       0 allocs/op
// Benchmark/RSTChanged-0637-0100-4     	   76028	     14424 ns/op	    8192 B/op	       0 allocs/op
func Benchmark(b *testing.B) {
	var str []rune
	for ti := range txts {
		if len(str) < len(txts[ti]) {
			str = txts[ti]
		}
	}
	var width uint
	for wi := range widths {
		if width < widths[wi] {
			width = widths[wi]
		}
	}
	drawer := func(row, col uint, r rune) {}
	cursor := func(row, col uint) {}
	name := fmt.Sprintf("%04d-%04d", len(str), width)
	ta := TextField{text: str}
	ta.SetWidth(width)
	ta.Render(drawer, cursor) // first step
	b.Run("Render-"+name, func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ta.Render(drawer, cursor)
		}
	})
	b.Run("Width-"+name, func(b *testing.B) {
		w := width
		var sw bool
		for n := 0; n < b.N; n++ {
			ta.SetWidth(w)
			if sw {
				w = w + 5
			} else {
				w = w - 5
			}
			sw = !sw
		}
	})
	b.Run("RWNoChange-"+name, func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ta.SetWidth(width)
			ta.Render(drawer, cursor)
		}
	})
	b.Run("RWChanged-"+name, func(b *testing.B) {
		w := width
		var sw bool
		for n := 0; n < b.N; n++ {
			ta.SetWidth(w)
			if sw {
				w = w + 5
			} else {
				w = w - 5
			}
			sw = !sw
			ta.Render(drawer, cursor)
		}
	})
	b.Run("RSTNoChange-"+name, func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			ta.SetText(str)
			ta.Render(drawer, cursor)
		}
	})
	b.Run("RSTChanged-"+name, func(b *testing.B) {
		var text []rune
		max := str
		min := str[:len(str)-4]
		var sw bool
		for n := 0; n < b.N; n++ {
			if sw {
				text = min
			} else {
				text = max
			}
			ta.SetText(text)
			sw = !sw
			ta.Render(drawer, cursor)
		}
	})
}
