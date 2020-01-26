package govaluate

/*
	Represents a single parsed token.
	代表解析后的片段
*/
type ExpressionToken struct {
	Kind  TokenKind // 类型
	Value interface{} // 值
}
