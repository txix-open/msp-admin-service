package sudir

import (
	"fmt"
)

type SudirAuthError struct {
	ErrorName        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (s *SudirAuthError) Error() string {
	return fmt.Sprintf("error: %s; description: %s", s.ErrorName, s.ErrorDescription)
}
