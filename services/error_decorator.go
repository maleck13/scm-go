package services

import (
	"runtime/debug"
)

//todo make more use of this

type DecoratedError struct {
	originalError error
	stack         []byte
	fullMsg       string
}

func (de *DecoratedError) Error() string {
	return de.fullMsg
}

func decorateError(additionalDetail string, err error) error {
	return &DecoratedError{err, debug.Stack(), "error detail " + additionalDetail + " original error " + err.Error()}
}
