package layer

import (
	"github.com/nbio/st"
	"testing"
)

func TestStack(t *testing.T) {
	s := &Stack{}
	first := MiddlewareFunc(nil)
	head := MiddlewareFunc(nil)
	tail := MiddlewareFunc(nil)

	s.Push(Normal, first)
	st.Expect(t, s.Len(), 1)
	s.Push(Head, head)
	st.Expect(t, s.Len(), 2)
	s.Push(Tail, tail)
	st.Expect(t, s.Len(), 3)

	st.Expect(t, s.Join()[0], head)
	st.Expect(t, s.Join()[1], first)
	st.Expect(t, s.Join()[2], tail)
}

func TestStackMemoization(t *testing.T) {
	s := &Stack{}

	first := MiddlewareFunc(nil)
	head := MiddlewareFunc(nil)
	tail := MiddlewareFunc(nil)

	s.Push(Normal, first)
	st.Expect(t, s.Len(), 1)
	s.Push(Head, head)
	st.Expect(t, s.Len(), 2)
	s.Push(Tail, tail)
	st.Expect(t, s.Len(), 3)

	memo := s.Join()
	st.Expect(t, memo[0], head)
	st.Expect(t, memo[1], first)
	st.Expect(t, memo[2], tail)
	st.Expect(t, s.memo, memo)

	s.Push(Tail, tail)
	st.Expect(t, s.Len(), 4)
	st.Expect(t, len(s.memo), 0)

	newMemo := s.Join()
	st.Expect(t, s.memo, newMemo)
	st.Expect(t, s.memo, s.Join())
}
