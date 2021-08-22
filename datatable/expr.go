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
	"errors"
	"github.com/BlueStorm001/gsql/util"
	"strings"
)

type tokenKind uint

const (
	valueKind tokenKind = iota
	symbolKind
)

type token struct {
	buf   []byte
	value string
	kind  tokenKind
}

type Wheres struct {
	Lh interface{}
	Op string
	Rh interface{}
	R  interface{}
}

type Expr struct {
	Tokens    []*token
	WhereExpr *Wheres
	OrderExpr []*Orders
	GroupExpr []*Groups
	next      int
	length    int
}

type Orders struct {
	Name string
	Op   string
}

type Groups struct {
	Name string
}

func GroupBy(buf []byte) *Expr {
	tokens := lex(buf)
	exp := &Expr{Tokens: tokens, length: len(tokens)}
	current := Groups{}
	for _, tok := range tokens {
		tok.value = util.BytToStr(tok.buf)
		switch tok.kind {
		case symbolKind:
			curCopy := current
			exp.GroupExpr = append(exp.GroupExpr, &curCopy)
			current = Groups{}
		case valueKind:
			current.Name = tok.value
		}
	}
	if current.Name != "" {
		curCopy := current
		exp.GroupExpr = append(exp.GroupExpr, &curCopy)
	}
	return exp
}

func OrderBy(buf []byte) *Expr {
	tokens := lex(buf)
	exp := &Expr{Tokens: tokens, length: len(tokens)}
	current := Orders{}
	for _, tok := range tokens {
		tok.value = util.BytToStr(tok.buf)
		switch tok.kind {
		case symbolKind:
			curCopy := current
			exp.OrderExpr = append(exp.OrderExpr, &curCopy)
			current = Orders{}
		case valueKind:
			value := strings.ToLower(tok.value)
			switch value {
			case "asc", "desc":
				current.Op = value
			default: //string
				current.Name = tok.value
			}
		}
	}
	if current.Name != "" {
		curCopy := current
		exp.OrderExpr = append(exp.OrderExpr, &curCopy)
	}
	return exp
}

func Where(buf []byte) (*Expr, error) {
	tokens := lex(buf)
	exp := &Expr{Tokens: tokens, length: len(tokens)}
	stmts := exp.stmt()
	if stmts.Rh == nil {
		switch v := stmts.Lh.(type) {
		case *Wheres:
			stmts = v
		}
	}
	if stmts.Lh == nil || stmts.Rh == nil || stmts.Op == "" {
		return exp, errors.New("stmt nil")
	}
	exp.WhereExpr = stmts
	return exp, nil
}

func (exp *Expr) stmt() *Wheres {
	if exp.next >= exp.length {
		return nil
	}
	root := &Wheres{}
	current := &Wheres{}
loop:
	for i := exp.next; i < exp.length; i++ {
		tok := exp.Tokens[i]
		tok.value = util.BytToStr(tok.buf)
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
				current = &Wheres{}
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

func (s *Wheres) lr(value interface{}, op ...string) *Wheres {
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
	case *Wheres:
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
		stmt := &Wheres{Lh: s}
		if newOp {
			stmt.Op = op[0]
		}
		if useValue {
			stmt.Rh = value
		}
		return stmt
	}
	if newOp {
		return &Wheres{Lh: s, Op: op[0]}
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
		case '\'', '"':
			if do {
				if escape {
					escape = false
					current.buf = append(current.buf, c)
				} else {
					if len(current.buf) == 0 {
						current.buf = append(current.buf, "''"...)
					}

					do = false
				}
			} else {
				do = true
			}
			continue
		case ' ', '=', '>', '<', '!', '(', ')', ',':
			if do {
				current.buf = append(current.buf, c)
				continue
			}
			if current.buf != nil {
				curCopy := current
				tokens = append(tokens, &curCopy)
			}
			if c != ' ' {
				tokens = append(tokens, &token{
					buf:  []byte{c},
					kind: symbolKind,
				})
			}
			current = token{}
		default:
			current.buf = append(current.buf, c)
		}
	}
	if current.buf != nil {
		tokens = append(tokens, &current)
	}
	return
}
