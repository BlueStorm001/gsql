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
	"strings"
)

type tokenKind uint

const (
	valueKind tokenKind = iota
	symbolKind
)

type token struct {
	value string
	kind  tokenKind
}

func (t *token) equals(other *token) bool {
	return t.value == other.value && t.kind == other.kind
}

func (t *token) finalize() bool {
	return true
}

func Parse(buf []byte) *Expr {
	tokens := lex(buf)
	exp := &Expr{Tokens: tokens, length: len(tokens)}
	exp.Tree = exp.stmt()
	return exp
}

type Stmt struct {
	Lh interface{}
	Op string
	Rh interface{}
}

type Expr struct {
	Tokens []*token
	Tree   *Stmt
	next   int
	length int
}

func (exp *Expr) stmt() *Stmt {
	if exp.next >= exp.length {
		return nil
	}
	root := &Stmt{}
	current := Stmt{}
loop:
	for i := exp.next; i < exp.length; i++ {
		tok := exp.Tokens[i]
		switch tok.kind {
		case symbolKind:
			switch tok.value {
			case "(":
				exp.next = i + 1
				s := exp.stmt()
				i = exp.next
				rs := root.lr(s)
				if rs != nil {
					root = rs
				}
			case ")":
				exp.next = i
				break loop
			default:
				current.Op += tok.value
			}
		case valueKind:
			value := strings.ToLower(tok.value)
			switch value {
			case "or", "and":
				rs := root.lr(current, value)
				if rs != nil {
					root = rs
				}
				current = Stmt{}
			default: //string
				current.lr(tok.value)
			}
		}
	}
	if current.Rh != nil || current.Lh != nil {
		root.lr(current)
	}
	return root
}

func (s *Stmt) lr(value interface{}, op ...string) *Stmt {
	var newOp bool
	if len(op) > 0 {
		if s.Op == "" {
			s.Op = op[0]
		} else {
			newOp = true
		}
	}
	var useValue = true
	switch v := value.(type) {
	case Stmt:
		if v.Lh == nil && v.Rh == nil {
			useValue = false
		}
	}
	if s.Lh == nil {
		if useValue {
			s.Lh = value
		}
	} else if s.Rh == nil {
		if useValue {
			s.Rh = value
		}
	} else {
		stmt := &Stmt{Lh: s}
		if newOp {
			stmt.Op = op[0]
		}
		if useValue {
			stmt.Rh = value
		}
		return stmt
	}
	if newOp {
		return &Stmt{Lh: s, Op: op[0]}
	}
	return nil
}

func lex(buf []byte) (tokens []*token) {
	current := token{}
	var do bool
	var escape bool
	for _, c := range buf {
		switch c {
		case '\n':
			continue
		case '\\':
			if do {
				escape = true
			}
			continue
		case '\'':
			if do {
				if escape {
					escape = false
					current.value += string(c)
				} else {
					do = false
				}
			} else {
				do = true
			}
			continue
		case ' ', '=', '>', '<', '!', '(', ')':
			if do {
				current.value += string(c)
				continue
			}
			if current.value != "" {
				curCopy := current
				tokens = append(tokens, &curCopy)
			}
			if c != ' ' {
				tokens = append(tokens, &token{
					value: string(c),
					kind:  symbolKind,
				})
			}
			current = token{}
		default:
			current.value += string(c)
		}
	}
	if current.value != "" {
		tokens = append(tokens, &current)
	}
	return
}
