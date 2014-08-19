package bind

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	StrQuote = '\''
)

type TokenType int

const (
	ExprToken TokenType = iota
	PuncToken
)

type ExprType int

const (
	ValueExpr ExprType = iota
	CallExpr
)

type token struct {
	kind TokenType
	v    string
}

type expr struct {
	name string
	typ  ExprType
	args []*expr
}

func isValidExprChar(c rune) bool {
	return c == '\'' || c == '.' || c == '_' || unicode.IsLetter(c) || unicode.IsDigit(c)
}

// tokenize simply splits the bind target string syntax into expressions (SomeObject.SomeField) and punctuations (().,), making
// it a little bit easier to parse
func tokenize(spec string) (tokens []token, err error) {
	tokens = make([]token, 0)
	err = nil
	var tok string
	flush := func() {
		if tok != "" {
			if strings.HasPrefix(tok, ".") || strings.HasSuffix(tok, ".") {
				err = errors.New("Invalid '.'")
				return
			}
			tokens = append(tokens, token{ExprToken, tok})
		}
		tok = ""
	}
	strlitMode := false //string literal mode
	for _, c := range spec {
		if !strlitMode {
			switch c {
			case ' ':
				if tok != "" {
					err = errors.New("Invalid space")
					return
				}
			case '(', ')', ',':
				flush()
				tokens = append(tokens, token{PuncToken, string(c)})
			case StrQuote:
				strlitMode = true
				tok += string(c)
			default:
				if isValidExprChar(c) {
					tok += string(c)
				} else {
					err = fmt.Errorf("Character '%q' is not allowed", c)
					return
				}
			}
		} else {
			if c == StrQuote {
				strlitMode = false
			} else if !unicode.IsDigit(c) && !unicode.IsLetter(c) && !strings.ContainsRune(",(-_.)", c) {
				err = fmt.Errorf("Use of characters other than numbers, " +
					"letters, parentheses ('(', ')'), dash ('-'), comma (','), " +
					"underscore ('_'), and dot ('.') is forbidden " +
					"inside string literals of bind string, " +
					"heavy processing and logic should not be in html template. Consider " +
					"moving your data to the model instead of putting it into the bind string.")
				return
			}
			tok += string(c)
		}
	}
	flush()

	return
}

// parse parses the bind target string, populate information into a tree of Expr pointers.
// Each helper call has a list arguments, each argument may be another helper call or an object expression.
func parse(spec string) (root *expr, err error) {
	tokens, err := tokenize(spec)
	if err != nil {
		return
	}
	invalid := func() {
		err = errors.New("Invalid syntax")
	}
	if len(tokens) == 0 {
		err = errors.New("Empty bind string")
	}
	if tokens[0].kind != ExprToken {
		invalid()
		return
	}
	stack := make([]*expr, 0)
	exprOf := make([]*expr, len(tokens))
	root = &expr{
		name: tokens[0].v,
		typ:  ValueExpr,
		args: make([]*expr, 0),
	}
	exprOf[0] = root
	var parent *expr = nil
	for ii, token := range tokens[1:] {
		i := ii + 1 //i starts from 1 instead of 1, more intuitive
		switch token.v {
		case "(":
			if tokens[i-1].kind != ExprToken {
				invalid()
				return
			}
			parent = exprOf[i-1]
			parent.typ = CallExpr
			stack = append(stack, parent)
		case ")":
			if parent == nil {
				invalid()
				return
			}
			stack = stack[:len(stack)-1]

		case ",":
			if !(tokens[i-1].kind == ExprToken || tokens[i-1].v == ")") {
				invalid()
				return
			}
		//expression
		default:
			e := &expr{
				name: tokens[i].v,
				typ:  ValueExpr,
				args: make([]*expr, 0),
			}
			exprOf[i] = e
			if len(stack) == 0 {
				invalid()
				return
			}
			stack[len(stack)-1].args = append(stack[len(stack)-1].args, e)
		}
	}

	return
}

func parseExpr(expr string) (value interface{}, isLiteral bool, err error) {
	err = nil
	isLiteral = true
	expr = strings.TrimSpace(expr)
	if expr == "true" || expr == "false" {
		value = (expr == "true")
		return
	}
	re := []rune(expr)
	numberMode := false
	floatMode := false
	for i, c := range expr {
		switch {
		case c == StrQuote:
			if i == 0 { //string literal
				if re[len(expr)-1] == StrQuote {
					value = string(re[1 : len(re)-1])
					return
				}
				err = fmt.Errorf("No matching quote.")
				return
			} else {
				err = fmt.Errorf("Invalid quote")
				return
			}
		case unicode.IsDigit(c):
			if i == 0 {
				numberMode = true
			}
		case unicode.IsLetter(c) || c == '_':
			if numberMode {
				err = fmt.Errorf("Invalid: dynamic expression cannot start with a number")
				return
			}
		case c == '.':
			if floatMode {
				err = fmt.Errorf("Multiple dot '.' for a number, invalid")
				return
			}
			if numberMode {
				floatMode = true
			}
		default:
			err = fmt.Errorf("Invalid character '%q'", c)
			return
		}
	}

	switch {
	case floatMode:
		var f float64
		f, err = strconv.ParseFloat(expr, 32)
		value = float32(f)
		return
	case numberMode:
		var i int
		i, err = strconv.Atoi(expr)
		value = i
		return
	default:
		isLiteral = false
		value = nil
		return
	}

	return
}
