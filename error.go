package rosie

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TypeError ...
func TypeError(exp reflect.Kind, got interface{}) error {
	return fmt.Errorf("wrong input type, expected %s but got %T", exp, got)
}

// Error ...
type Error struct {
	TaskName string
	Err      error
}

// Error implements error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.TaskName, e.Err.Error())
}

// InitError can be recovered from a panic fired by Cmd function.
type InitError struct {
	msg string
	err error
}

// Error implements error interface.
func (e *InitError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s", e.msg, e.err)
	}
	return e.msg
}

// MultiError ...
type MultiError struct {
	Err []error
}

func appendError(left, right error) error {
	switch {
	case left == nil:
		return right
	case right == nil:
		return left
	}

	_, leftOK := left.(*MultiError)
	_, rightOK := right.(*MultiError)

	var n MultiError
	if leftOK {
		n.Err = append(n.Err, left.(*MultiError).Err...)
	}
	if rightOK {
		n.Err = append(n.Err, right.(*MultiError).Err...)
	}
	if len(n.Err) == 0 {
		n.Err = append(n.Err, left, right)
	}

	return &n
}

/* #nosec */
func (e *MultiError) Error() string {
	var sb strings.Builder
	sb.WriteString("rosie: multi-error:")

	if len(e.Err) == 0 {
		sb.WriteString(" no errors")
	}

	for i, err := range e.Err {
		sb.WriteRune('\n')
		sb.WriteString(strconv.FormatInt(int64(i), 64))
		sb.WriteRune(':')
		sb.WriteRune('	')
		sb.WriteString(err.Error())
	}

	return sb.String()
}
