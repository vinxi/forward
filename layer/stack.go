package layer

// Priority represents the middleware priority.
type Priority int

const (
	// TopHead priority defines the middleware handler
	// as first element of the stack head.
	TopHead Priority = iota

	// Head priority stores the middleware handler
	// in the head of the stack.
	Head

	// Normal priority defines the middleware handler
	// in the last stack available.
	Normal

	// TopTail priority defines the middleware handler
	// as fist element of the stack tail.
	TopTail

	// Tail priority defines the middleware handler
	// in the tail of the stack.
	Tail
)

// Stack stores the data to show.
type Stack struct {
	// memo stores the memorized pre-computed merged stack for better performance.
	memo []MiddlewareFunc

	// Head stores the head priority handlers.
	Head []MiddlewareFunc

	// Stack stores the middleware normal priority handlers.
	Stack []MiddlewareFunc

	// Tail stores the middleware tail priority handlers.
	Tail []MiddlewareFunc
}

// Push pushes a new middleware handler to the stack based on the given priority.
func (s *Stack) Push(order Priority, h MiddlewareFunc) {
	s.memo = nil // flush the memoized stack
	if order == TopHead {
		s.Head = append([]MiddlewareFunc{h}, s.Head...)
	}
	if order == Head {
		s.Head = append(s.Head, h)
	}
	if order == Tail {
		s.Tail = append(s.Tail, h)
	}
	if order == TopTail {
		s.Tail = append([]MiddlewareFunc{h}, s.Tail...)
	}
	if order == Normal {
		s.Stack = append(s.Stack, h)
	}
}

// Join joins the middleware functions into a unique slice.
func (s *Stack) Join() []MiddlewareFunc {
	if s.memo != nil {
		return s.memo
	}
	s.memo = append(append(s.Head, s.Stack...), s.Tail...)
	return s.memo
}

// Len returns the middleware stack length.
func (s *Stack) Len() int {
	return len(s.Stack) + len(s.Tail) + len(s.Head)
}
