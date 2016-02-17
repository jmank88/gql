package errors

import (
	"fmt"
)

type SyntaxError struct {
	// Position in source. Rune offset.
	Pos int
	Err error
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("Syntax error at position %d: %s", e.Pos, e.Err)
}
