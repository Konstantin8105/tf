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

func (b Buffer) HasError() bool {
	for r := range b.m {
		for c := range b.m[r] {
			if b.m[r][c] == errorRune {
				return true
			}
		}
	}
	return false
}

const testdata = "testdata"

func Test(t *testing.T) {
	f := String
	for ti := range txts {
		for wi := range widths {
			// static
			name := fmt.Sprintf("%04d-%04d", len(txts[ti]), widths[wi])
			basename := name
			ta := TextArea{
				Text:   txts[ti],
				Format: f,
			}
			t.Run(name, func(t *testing.T) {
				ta.SetWidth(widths[wi])
				snapshot(t, name, ta)
			})
			// resize
			if wi == 0 {
				continue
			}
			name += fmt.Sprintf("->%04d", widths[wi-1])
			t.Run(name, func(t *testing.T) {
				ta.SetWidth(widths[wi-1])
				snapshot(t, name, ta)
			})
			// return
			name += fmt.Sprintf("->%04d", widths[wi])
			t.Run(name, func(t *testing.T) {
				ta.SetWidth(widths[wi])
				snapshot(t, name, ta)
			})
			// compare
			t.Run(name+"-compare", func(t *testing.T) {
				compare(t, basename, name)
			})
			// cursor move
			name = basename
			mv := []func(){
				ta.CursorMoveUp,
				ta.CursorMoveDown,
				ta.CursorMoveLeft,
				ta.CursorMoveRight,
				ta.CursorMoveHome,
				ta.CursorMoveEnd,
				ta.CursorPageDown,
				ta.CursorPageUp,
			}
			for im := range mv {
				name := fmt.Sprintf("%s-cursor%03d", name, im)
				for times := 0; times < 10; times++ {
					name := fmt.Sprintf("%s-times%02d", name, times)
					t.Run(name , func(t *testing.T) {
						mv[im]()
						snapshot(t, name, ta)
					})
				}
			}
		}
	}
}

func repeat(number int, f func()) {
	for i := 0; i < number; i++ {
		f()
	}
}

func snapshot(t *testing.T, name string, ta TextArea) {
	var b Buffer
	ta.Render(b.Drawer)
	actual := fmt.Sprintf("%s", b)
	if b.HasError() {
		t.Fatalf("buffer have error rune")
	}
	// for update test screens run in console:
	// UPDATE=true go test
	filename := filepath.Join(testdata, name)
	if os.Getenv("UPDATE") == "true" {
		if err := ioutil.WriteFile(filename, []byte(actual), 0644); err != nil {
			t.Fatalf("Cannot write snapshot to file: %v", err)
		}
	}
	// get expect result
	expect, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Cannot read snapshot file: %v", err)
	}
	// compare
	if !bytes.Equal([]byte(actual), expect) {
		t.Errorf("Snapshots is not same:\n%s\n%s", expect, string(actual))
	}
}

func compare(t *testing.T, name1, name2 string) {
	if name1 == name2 {
		t.Fatalf("compare test names are same")
	}
	filename1 := filepath.Join(testdata, name1)
	expect1, err := ioutil.ReadFile(filename1)
	if err != nil {
		t.Fatalf("Cannot read snapshot file: %v", err)
	}
	filename2 := filepath.Join(testdata, name2)
	expect2, err := ioutil.ReadFile(filename2)
	if err != nil {
		t.Fatalf("Cannot read snapshot file: %v", err)
	}
	// compare
	if !bytes.Equal(expect1, expect2) {
		t.Errorf("Comparing is not same:\n%s\n%s", expect1, expect2)
	}
}
