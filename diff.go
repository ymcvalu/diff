package diff

import (
	"bytes"
	"fmt"
	"io"
)

// A type that satisfies Diffable can be diffed using Myers.
// A is the initial state; B is the final state.
type Diffable interface {
	LenA() int
	LenB() int
	Equal(ai, bi int) bool
}

// A type that is Writeable can be written using WriteUnified.
// A is the initial state; B is the final state.
type Writeable interface {
	WriteToA(w io.Writer, i int) (int, error)
	WriteToB(w io.Writer, i int) (int, error)
	FilenameA() string
	FilenameB() string
}

// DiffWriteable is the union of Diffable and Writeable.
// TODO: better name
type DiffWriteable interface {
	Diffable
	Writeable
}

// TODO: consider adding a StringIntern type, something like:
//
// type StringIntern struct {
// 	s map[string]*string
// }
//
// func (i *StringIntern) Bytes(b []byte) *string
// func (i *StringIntern) String(s string) *string
//
// And document what it is and why to use it.
// And consider adding helper functions to Strings and Bytes to use it.
// The reason to use it is that a lot of the execution time in diffing
// (which is an expensive operation) is taken up doing string comparisons.
// If you have paid the O(n) cost to intern all strings involved in both A and B,
// then string comparisons are reduced to cheap pointer comparisons.

// An op is a edit operation used to transform A into B.
type op int8

//go:generate stringer -type op

const (
	del op = -1
	eq  op = 0
	ins op = 1
)

// A segment is a set of steps of the same op.
type segment struct {
	FromA, ToA int // Beginning and ending indices into A of this operation
	FromB, ToB int // ditto, for B
}

func (s segment) op() op {
	if s.FromA == s.ToA {
		return ins
	}
	if s.FromB == s.ToB {
		return del
	}
	return eq
}

func (s segment) String() string {
	// This output is helpful when hacking on a Myers diff.
	// In other contexts it is usually more natural to group FromA, ToA and FromB, ToB.
	return fmt.Sprintf("(%d, %d) -- %s %d --> (%d, %d)", s.FromA, s.FromB, s.op(), s.Len(), s.ToA, s.ToB)
}

func (s segment) Len() int {
	if s.FromA == s.ToA {
		return s.ToB - s.FromB
	}
	return s.ToA - s.FromA
}

// An EditScript is an edit script to alter A into B.
type EditScript struct {
	segs []segment
}

// IsIdentity reports whether e is the identity edit script, that is, whether A and B are identical.
// See the TestHelper example.
func (e *EditScript) IsIdentity() bool {
	for _, seg := range e.segs {
		if seg.op() != eq {
			return false
		}
	}
	return true
}

// TODO: consider adding an "it just works" test helper that accepts two slices (via interface{}),
// diffs them using Strings or Bytes or Slices (using reflect.DeepEqual) as appropriate,
// and calls t.Errorf with a generated diff if they're not equal.

// scriptWithSegments returns an EditScript containing s.
// It is used to reduce line noise.
func scriptWithSegments(s ...segment) EditScript {
	return EditScript{segs: s}
}

// dump formats s for debugging.
func (s EditScript) dump() string {
	buf := new(bytes.Buffer)
	for _, seg := range s.segs {
		fmt.Fprintln(buf, seg)
	}
	return buf.String()
}
