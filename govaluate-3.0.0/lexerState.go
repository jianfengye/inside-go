package govaluate

import (
	"errors"
	"fmt"
)

// 这个结构用于表示每个分析器分析出来一个token之后的状态
type lexerState struct {
	isEOF          bool        // 是否是已经终结了，比如token是左括号，就没有终结
	isNullable     bool        // 是否是可null
	kind           TokenKind   // 当前token的类型
	validNextKinds []TokenKind // 后续token的允许类型
}

// lexer states.
// Constant for all purposes except compiler.
var validLexerStates = []lexerState{

	lexerState{
		kind:       UNKNOWN,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			PATTERN,
			FUNCTION,
			STRING,
			TIME,
			CLAUSE,
		},
	},

	lexerState{

		kind:       CLAUSE,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			PATTERN,
			FUNCTION,
			STRING,
			TIME,
			CLAUSE,
			CLAUSE_CLOSE,
		},
	},

	lexerState{

		kind:       CLAUSE_CLOSE,
		isEOF:      true,
		isNullable: true,
		validNextKinds: []TokenKind{

			COMPARATOR,
			MODIFIER,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			STRING,
			PATTERN,
			TIME,
			CLAUSE,
			CLAUSE_CLOSE,
			LOGICALOP,
			TERNARY,
			SEPARATOR,
		},
	},

	lexerState{

		kind:       NUMERIC,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	lexerState{

		kind:       BOOLEAN,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	lexerState{

		kind:       STRING,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	lexerState{

		kind:       TIME,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			SEPARATOR,
		},
	},
	lexerState{

		kind:       PATTERN,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			SEPARATOR,
		},
	},
	lexerState{

		kind:       VARIABLE,
		isEOF:      true,
		isNullable: false,
		validNextKinds: []TokenKind{

			MODIFIER,
			COMPARATOR,
			LOGICALOP,
			CLAUSE_CLOSE,
			TERNARY,
			SEPARATOR,
		},
	},
	lexerState{

		kind:       MODIFIER,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			VARIABLE,
			FUNCTION,
			STRING,
			BOOLEAN,
			CLAUSE,
			CLAUSE_CLOSE,
		},
	},
	lexerState{

		kind:       COMPARATOR,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			FUNCTION,
			STRING,
			TIME,
			CLAUSE,
			CLAUSE_CLOSE,
			PATTERN,
		},
	},
	lexerState{

		kind:       LOGICALOP,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			VARIABLE,
			FUNCTION,
			STRING,
			TIME,
			CLAUSE,
			CLAUSE_CLOSE,
		},
	},
	lexerState{

		kind:       PREFIX,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			NUMERIC,
			BOOLEAN,
			VARIABLE,
			FUNCTION,
			CLAUSE,
			CLAUSE_CLOSE,
		},
	},

	lexerState{

		kind:       TERNARY,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			STRING,
			TIME,
			VARIABLE,
			FUNCTION,
			CLAUSE,
			SEPARATOR,
		},
	},
	lexerState{

		kind:       FUNCTION,
		isEOF:      false,
		isNullable: false,
		validNextKinds: []TokenKind{
			CLAUSE,
		},
	},
	lexerState{

		kind:       SEPARATOR,
		isEOF:      false,
		isNullable: true,
		validNextKinds: []TokenKind{

			PREFIX,
			NUMERIC,
			BOOLEAN,
			STRING,
			TIME,
			VARIABLE,
			FUNCTION,
			CLAUSE,
		},
	},
}

func (this lexerState) canTransitionTo(kind TokenKind) bool {

	for _, validKind := range this.validNextKinds {

		if validKind == kind {
			return true
		}
	}

	return false
}

func checkExpressionSyntax(tokens []ExpressionToken) error {

	var state lexerState
	var lastToken ExpressionToken
	var err error

	state = validLexerStates[0]

	for _, token := range tokens {

		if !state.canTransitionTo(token.Kind) {

			// call out a specific error for tokens looking like they want to be functions.
			if lastToken.Kind == VARIABLE && token.Kind == CLAUSE {
				return errors.New("Undefined function " + lastToken.Value.(string))
			}

			firstStateName := fmt.Sprintf("%s [%v]", state.kind.String(), lastToken.Value)
			nextStateName := fmt.Sprintf("%s [%v]", token.Kind.String(), token.Value)

			return errors.New("Cannot transition token types from " + firstStateName + " to " + nextStateName)
		}

		state, err = getLexerStateForToken(token.Kind)
		if err != nil {
			return err
		}

		if !state.isNullable && token.Value == nil {

			errorMsg := fmt.Sprintf("Token kind '%v' cannot have a nil value", token.Kind.String())
			return errors.New(errorMsg)
		}

		lastToken = token
	}

	if !state.isEOF {
		return errors.New("Unexpected end of expression")
	}
	return nil
}

func getLexerStateForToken(kind TokenKind) (lexerState, error) {

	for _, possibleState := range validLexerStates {

		if possibleState.kind == kind {
			return possibleState, nil
		}
	}

	errorMsg := fmt.Sprintf("No lexer state found for token kind '%v'\n", kind.String())
	return validLexerStates[0], errors.New(errorMsg)
}
