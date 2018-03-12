package xlog

import "fmt"

type ErrConfigObject struct {
	expect string
	got    interface{}
}

func (err ErrConfigObject) Error() string {
	return fmt.Sprintf("config object is not an instance of %s, instead got '%T'", err.expect, err.got)
}

type ErrInvalidLevel struct{}

func (err ErrInvalidLevel) Error() string {
	return "input level is not one of: TRACE, INFO, WARN, ERROR or FATAL"
}

type ErrMissing struct{}

func (err ErrMissing) Error() string {
	return "(Missing)"
}
