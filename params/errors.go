package params

import (
	"fmt"
	"net/http"
)

type enumInvalidValueError struct {
	value string
}

func (e *enumInvalidValueError) Error() string {
	return fmt.Sprintf("%q is not a valid enumerated value", e.value)
}

func (e *enumInvalidValueError) StatusCode() int {
	return http.StatusUnprocessableEntity
}
