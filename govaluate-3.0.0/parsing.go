package govaluate

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode"
)

// 这是解析token最核心的函数。
// 将表达式拆分，如果有自定义函数，自定义的函数也按照特定的方法进行拆分
func parseTokens(expression string, functions map[string]ExpressionFunction) ([]ExpressionToken, error) {

	var ret []ExpressionToken
	var token ExpressionToken
	var stream *lexerStream
	var state lexerState
	var err error
	var found bool

	// 将字符串转变为 lexerStream 结构
	stream = newLexerStream(expression)
	// 将初始状态设置为unknown
	state = validLexerStates[0]

	// 读取一个token，变更状态
	for stream.canRead() {

		token, err, found = readToken(stream, state, functions)

		if err != nil {
			return ret, err
		}

		if !found {
			break
		}

		state, err = getLexerStateForToken(token.Kind)
		if err != nil {
			return ret, err
		}

		// append this valid token
		ret = append(ret, token)
	}

	// 最终读取完成之后的tokens检查一下是否平衡（有左括号，是否有右括号）
	err = checkBalance(ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// 读取一个token, 参数是stream-当前读取的表达式，state-当前读取的token状态，function-自定义函数
func readToken(stream *lexerStream, state lexerState, functions map[string]ExpressionFunction) (ExpressionToken, error, bool) {

	var function ExpressionFunction
	var ret ExpressionToken
	var tokenValue interface{}
	var tokenTime time.Time
	var tokenString string
	var kind TokenKind
	var character rune
	var found bool
	var completed bool
	var err error

	// numeric is 0-9, or .
	// string starts with '
	// variable is alphanumeric, always starts with a letter
	// bracket always means variable
	// symbols are anything non-alphanumeric
	// all others read into a buffer until they reach the end of the stream
	for stream.canRead() {

		character = stream.readCharacter()

		if unicode.IsSpace(character) {
			continue
		}

		kind = UNKNOWN

		// numeric constant
		if isNumeric(character) {

			tokenString = readTokenUntilFalse(stream, isNumeric)
			tokenValue, err = strconv.ParseFloat(tokenString, 64)

			if err != nil {
				errorMsg := fmt.Sprintf("Unable to parse numeric value '%v' to float64\n", tokenString)
				return ExpressionToken{}, errors.New(errorMsg), false
			}
			kind = NUMERIC
			break
		}

		// comma, separator
		if character == ',' {

			tokenValue = ","
			kind = SEPARATOR
			break
		}

		// escaped variable
		if character == '[' {

			tokenValue, completed = readUntilFalse(stream, true, false, true, isNotClosingBracket)
			kind = VARIABLE

			if !completed {
				return ExpressionToken{}, errors.New("Unclosed parameter bracket"), false
			}

			// above method normally rewinds us to the closing bracket, which we want to skip.
			stream.rewind(-1)
			break
		}

		// regular variable - or function?
		if unicode.IsLetter(character) {

			tokenString = readTokenUntilFalse(stream, isVariableName)

			tokenValue = tokenString
			kind = VARIABLE

			// boolean?
			if tokenValue == "true" {

				kind = BOOLEAN
				tokenValue = true
			} else {

				if tokenValue == "false" {

					kind = BOOLEAN
					tokenValue = false
				}
			}

			// textual operator?
			if tokenValue == "in" || tokenValue == "IN" {

				// force lower case for consistency
				tokenValue = "in"
				kind = COMPARATOR
			}

			// function?
			function, found = functions[tokenString]
			if found {
				kind = FUNCTION
				tokenValue = function
			}
			break
		}

		if !isNotQuote(character) {
			tokenValue, completed = readUntilFalse(stream, true, false, true, isNotQuote)

			if !completed {
				return ExpressionToken{}, errors.New("Unclosed string literal"), false
			}

			// advance the stream one position, since reading until false assumes the terminator is a real token
			stream.rewind(-1)

			// check to see if this can be parsed as a time.
			tokenTime, found = tryParseTime(tokenValue.(string))
			if found {
				kind = TIME
				tokenValue = tokenTime
			} else {
				kind = STRING
			}
			break
		}

		if character == '(' {
			tokenValue = character
			kind = CLAUSE
			break
		}

		if character == ')' {
			tokenValue = character
			kind = CLAUSE_CLOSE
			break
		}

		// must be a known symbol
		tokenString = readTokenUntilFalse(stream, isNotAlphanumeric)
		tokenValue = tokenString

		// quick hack for the case where "-" can mean "prefixed negation" or "minus", which are used
		// very differently.
		if state.canTransitionTo(PREFIX) {
			_, found = prefixSymbols[tokenString]
			if found {

				kind = PREFIX
				break
			}
		}
		_, found = modifierSymbols[tokenString]
		if found {

			kind = MODIFIER
			break
		}

		_, found = logicalSymbols[tokenString]
		if found {

			kind = LOGICALOP
			break
		}

		_, found = comparatorSymbols[tokenString]
		if found {

			kind = COMPARATOR
			break
		}

		_, found = ternarySymbols[tokenString]
		if found {

			kind = TERNARY
			break
		}

		errorMessage := fmt.Sprintf("Invalid token: '%s'", tokenString)
		return ret, errors.New(errorMessage), false
	}

	ret.Kind = kind
	ret.Value = tokenValue

	return ret, nil, (kind != UNKNOWN)
}

func readTokenUntilFalse(stream *lexerStream, condition func(rune) bool) string {

	var ret string

	stream.rewind(1)
	ret, _ = readUntilFalse(stream, false, true, true, condition)
	return ret
}

/*
	Returns the string that was read until the given [condition] was false, or whitespace was broken.
	Returns false if the stream ended before whitespace was broken or condition was met.
*/
func readUntilFalse(stream *lexerStream, includeWhitespace bool, breakWhitespace bool, allowEscaping bool, condition func(rune) bool) (string, bool) {

	var tokenBuffer bytes.Buffer
	var character rune
	var conditioned bool

	conditioned = false

	for stream.canRead() {

		character = stream.readCharacter()

		// Use backslashes to escape anything
		if allowEscaping && character == '\\' {

			character = stream.readCharacter()
			tokenBuffer.WriteString(string(character))
			continue
		}

		if unicode.IsSpace(character) {

			if breakWhitespace && tokenBuffer.Len() > 0 {
				conditioned = true
				break
			}
			if !includeWhitespace {
				continue
			}
		}

		if condition(character) {
			tokenBuffer.WriteString(string(character))
		} else {
			conditioned = true
			stream.rewind(1)
			break
		}
	}

	return tokenBuffer.String(), conditioned
}

/*
	Checks to see if any optimizations can be performed on the given [tokens], which form a complete, valid expression.
	The returns slice will represent the optimized (or unmodified) list of tokens to use.
	使用一些规则做token的优化，其实只做一件事，判断是否是正则匹配表达式
*/
func optimizeTokens(tokens []ExpressionToken) ([]ExpressionToken, error) {

	var token ExpressionToken
	var symbol OperatorSymbol
	var err error
	var index int

	for index, token = range tokens {

		// if we find a regex operator, and the right-hand value is a constant, precompile and replace with a pattern.
		// 如果是 =~ 或者 !~ 且后面是一个跟着一个字符串，则是判断是否正则匹配。
		if token.Kind != COMPARATOR {
			continue
		}

		symbol = comparatorSymbols[token.Value.(string)]
		if symbol != REQ && symbol != NREQ {
			continue
		}

		index++
		token = tokens[index]
		if token.Kind == STRING {

			token.Kind = PATTERN
			token.Value, err = regexp.Compile(token.Value.(string))

			if err != nil {
				return tokens, err
			}

			tokens[index] = token
		}
	}
	return tokens, nil
}

/*
	Checks the balance of tokens which have multiple parts, such as parenthesis.
	判断做括号和右括号数字是否个数一样
*/
func checkBalance(tokens []ExpressionToken) error {

	var stream *tokenStream
	var token ExpressionToken
	var parens int

	stream = newTokenStream(tokens)

	for stream.hasNext() {

		token = stream.next()
		if token.Kind == CLAUSE {
			parens++
			continue
		}
		if token.Kind == CLAUSE_CLOSE {
			parens--
			continue
		}
	}

	if parens != 0 {
		return errors.New("Unbalanced parenthesis")
	}
	return nil
}

func isNumeric(character rune) bool {

	return unicode.IsDigit(character) || character == '.'
}

func isNotQuote(character rune) bool {

	return character != '\'' && character != '"'
}

func isNotAlphanumeric(character rune) bool {

	return !(unicode.IsDigit(character) ||
		unicode.IsLetter(character) ||
		character == '(' ||
		character == ')' ||
		!isNotQuote(character))
}

func isVariableName(character rune) bool {

	return unicode.IsLetter(character) ||
		unicode.IsDigit(character) ||
		character == '_'
}

func isNotClosingBracket(character rune) bool {

	return character != ']'
}

/*
	Attempts to parse the [candidate] as a Time.
	Tries a series of standardized date formats, returns the Time if one applies,
	otherwise returns false through the second return.
*/
func tryParseTime(candidate string) (time.Time, bool) {

	var ret time.Time
	var found bool

	timeFormats := [...]string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.Kitchen,
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",                         // RFC 3339
		"2006-01-02 15:04",                   // RFC 3339 with minutes
		"2006-01-02 15:04:05",                // RFC 3339 with seconds
		"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
		"2006-01-02T15Z0700",                 // ISO8601 with hour
		"2006-01-02T15:04Z0700",              // ISO8601 with minutes
		"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds
		"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds
	}

	for _, format := range timeFormats {

		ret, found = tryParseExactTime(candidate, format)
		if found {
			return ret, true
		}
	}

	return time.Now(), false
}

func tryParseExactTime(candidate string, format string) (time.Time, bool) {

	var ret time.Time
	var err error

	ret, err = time.ParseInLocation(format, candidate, time.Local)
	if err != nil {
		return time.Now(), false
	}

	return ret, true
}
