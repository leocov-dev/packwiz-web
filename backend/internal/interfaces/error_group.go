package interfaces

import (
	"strings"
)

type ErrorGroup struct {
	Errors []error
}

func NewErrorGroup() *ErrorGroup {
	return &ErrorGroup{}
}

// Add adds a new error to the group
func (eg *ErrorGroup) Add(err error) {
	if err != nil {
		eg.Errors = append(eg.Errors, err)
	}
}

// Error implements the error interface to combine all errors into a single string
func (eg *ErrorGroup) Error() string {
	if len(eg.Errors) == 0 {
		return ""
	}

	var msgs []string
	for _, err := range eg.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// IsEmpty checks if the error group is empty
func (eg *ErrorGroup) IsEmpty() bool {
	return len(eg.Errors) == 0
}
