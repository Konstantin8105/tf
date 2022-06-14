package tf

// text agnostic interface
type TextField struct {
	cursor struct {
		line   int
		offset int
	}
	Text   []rune
	Render []rune
	filter int // 0 - string; 1 - float; 2 - int
}

func (t *TextField) CursorMoveUp()    {}
func (t *TextField) CursorMoveDown()  {}
func (t *TextField) CursorMoveLeft()  {}
func (t *TextField) CursorMoveRight() {}
func (t *TextField) CursorMoveHome()  {}
func (t *TextField) CursorMoveEnd()   {}
func (t *TextField) CursorPageDown()  {}
func (t *TextField) CursorPageUp()    {}
func (t *TextField) DoubleClick()     {}
func (t *TextField) InsertRune()      {}
func (t *TextField) RemoveBackspace() {}
func (t *TextField) RemoveDel()       {}
