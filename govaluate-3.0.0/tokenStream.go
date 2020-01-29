package govaluate

// 这个结构是为了读取token串
type tokenStream struct {
	tokens      []ExpressionToken // token串
	index       int               // 读取位置
	tokenLength int               // token总个数
}

func newTokenStream(tokens []ExpressionToken) *tokenStream {

	var ret *tokenStream

	ret = new(tokenStream)
	ret.tokens = tokens
	ret.tokenLength = len(tokens)
	return ret
}

func (this *tokenStream) rewind() {
	this.index -= 1
}

func (this *tokenStream) next() ExpressionToken {

	var token ExpressionToken

	token = this.tokens[this.index]

	this.index += 1
	return token
}

func (this tokenStream) hasNext() bool {

	return this.index < this.tokenLength
}
