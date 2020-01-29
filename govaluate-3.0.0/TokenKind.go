package govaluate

/*
	Represents all valid types of tokens that a token can be.
	代表token有多少种类型
*/
type TokenKind int

const (
	UNKNOWN TokenKind = iota

	PREFIX    // 前缀 ! - ~
	NUMERIC   // 数字 12345.678
	BOOLEAN   // 布尔值 true false
	STRING    // 字符串 foobar
	PATTERN   // 正则
	TIME      // 时间
	VARIABLE  // 变量
	FUNCTION  // 函数
	SEPARATOR // 分隔符， 逗号

	COMPARATOR // 比较符号 > >= < <= == != =~ !~
	LOGICALOP  // 逻辑操作 || &&
	MODIFIER   // 修改器 + - / * & | ^ ** % >> <<

	CLAUSE
	CLAUSE_CLOSE

	TERNARY // 三元运算符 ? :
)

/*
	GetTokenKindString returns a string that describes the given TokenKind.
	e.g., when passed the NUMERIC TokenKind, this returns the string "NUMERIC".
*/
func (kind TokenKind) String() string {

	switch kind {

	case PREFIX:
		return "PREFIX"
	case NUMERIC:
		return "NUMERIC"
	case BOOLEAN:
		return "BOOLEAN"
	case STRING:
		return "STRING"
	case PATTERN:
		return "PATTERN"
	case TIME:
		return "TIME"
	case VARIABLE:
		return "VARIABLE"
	case FUNCTION:
		return "FUNCTION"
	case SEPARATOR:
		return "SEPARATOR"
	case COMPARATOR:
		return "COMPARATOR"
	case LOGICALOP:
		return "LOGICALOP"
	case MODIFIER:
		return "MODIFIER"
	case CLAUSE:
		return "CLAUSE"
	case CLAUSE_CLOSE:
		return "CLAUSE_CLOSE"
	case TERNARY:
		return "TERNARY"
	}

	return "UNKNOWN"
}
