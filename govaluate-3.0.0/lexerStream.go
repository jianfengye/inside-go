package govaluate

// 这个结构主要用于做字符串和rune串的转换
type lexerStream struct {
	source   []rune // 字符转换为rune
	position int    // 当前读取的rune位置
	length   int    // 总的rune长度
}

func newLexerStream(source string) *lexerStream {

	var ret *lexerStream
	var runes []rune

	for _, character := range source {
		runes = append(runes, character)
	}

	ret = new(lexerStream)
	ret.source = runes
	ret.length = len(runes)
	return ret
}

func (this *lexerStream) readCharacter() rune {

	var character rune

	character = this.source[this.position]
	this.position += 1
	return character
}

func (this *lexerStream) rewind(amount int) {
	this.position -= amount
}

func (this lexerStream) canRead() bool {
	return this.position < this.length
}
