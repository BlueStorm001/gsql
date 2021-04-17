// Copyright (c) 2021 BlueStorm
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFINGEMENT IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package datatable

import (
	"fmt"
)

type location struct {
	line uint
	col  uint
}
type keyword string

const (
	selectKeyword keyword = "or"
	fromKeyword   keyword = "and"
	asKeyword     keyword = "as"
	tableKeyword  keyword = "table"
	createKeyword keyword = "create"
	insertKeyword keyword = "insert"
	intoKeyword   keyword = "into"
	valuesKeyword keyword = "values"
	intKeyword    keyword = "int"
	textKeyword   keyword = "text"
)

type symbol string

const (
	semicolonSymbol  symbol = ";"
	asteriskSymbol   symbol = "*"
	commaSymbol      symbol = ","
	leftparenSymbol  symbol = "("
	rightparenSymbol symbol = ")"
)

type tokenKind uint

const (
	keywordKind tokenKind = iota
	symbolKind
	identifierKind
	stringKind
	numericKind
)

type token struct {
	value string
	op    string
	kind  tokenKind
	loc   location
}

func (t *token) equals(other *token) bool {
	return t.value == other.value && t.kind == other.kind
}

func (t *token) finalize() bool {
	return true
}

func Lex(buf []byte) ([]*token, error) {
	return lex(buf)
}

func lex(buf []byte) ([]*token, error) {
	var tokens []*token
	current := token{}
	var line uint = 0
	var col uint = 0
	for _, c := range buf {
		switch c {
		case '\n':
			line++
			col = 0
			continue
		case ' ':
			fallthrough
		case '=':
			fallthrough
		case '>':
			fallthrough
		case '<':
			fallthrough
		case '!':
			fallthrough
		case '(':
			fallthrough
		case ')':
			if !current.finalize() {
				return nil, fmt.Errorf("Unexpected token '%s' at %d:%d", current.value, current.loc.line, current.loc.col)
			}
			if current.value != "" {
				curCopy := current
				tokens = append(tokens, &curCopy)
			}
			switch c {
			case '(', ')', '=', '<', '>', '!':
				tokens = append(tokens, &token{
					loc:   location{col: col, line: line},
					value: string(c),
					kind:  symbolKind,
				})
			}
			current = token{}
			current.loc.col = col
			current.loc.line = line
		default:
			current.value += string(c)
		}
		col++
	}
	return tokens, nil
}
