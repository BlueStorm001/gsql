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

package util

import (
	"strconv"
	"unsafe"
)

type Builder struct {
	addr *Builder // of receiver, to detect copies by value
	buf  []byte
	mb   bool
}

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input. noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
// This was copied from the runtime; see issues 23382 and 7921.
//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func (builder *Builder) copyCheck() {
	if builder.addr == nil {
		builder.addr = (*Builder)(noescape(unsafe.Pointer(builder)))
	} else if builder.addr != builder {
		panic("strings: illegal use of non-zero Builder copied by value")
	}
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (builder *Builder) Append(str string) *Builder {
	builder.copyCheck()
	builder.buf = append(builder.buf, str...)
	return builder
}

func (builder *Builder) AppendInt(num int) *Builder {
	builder.copyCheck()
	builder.buf = append(builder.buf, strconv.Itoa(num)...)
	return builder
}

func (builder *Builder) AppendBytes(buf []byte) *Builder {
	builder.copyCheck()
	builder.buf = append(builder.buf, buf...)
	return builder
}

func (builder *Builder) AppendByte(b byte) *Builder {
	builder.copyCheck()
	builder.buf = append(builder.buf, b)
	return builder
}

func (builder *Builder) Mb() *Builder {
	builder.mb = true
	return builder
}

func (builder *Builder) String() string {
	return *(*string)(unsafe.Pointer(&builder.buf))
}

func (builder *Builder) ToString() string {
	return string(builder.buf)
}

func (builder *Builder) Bytes() []byte {
	return builder.buf
}

func (builder *Builder) Len() int {
	return len(builder.buf)
}

func (builder *Builder) Clear() {
	builder.addr = nil
	builder.buf = nil
	builder.mb = false
}

func (builder *Builder) Reset() {
	builder.buf = builder.buf[:0]
	builder.mb = false
}
