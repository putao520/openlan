//
package libol

import "fmt"

type Err struct {
	Code    int
	Message string
}

func Errer(message string, v ...interface{}) (e *Err) {
	e = &Err{
		Message: fmt.Sprintf(message, v...),
	}
	return
}

func (e *Err) String() string {
	return fmt.Sprint("%d: %s", e.Code, e.Message)
}

func (e *Err) Error() string {
	return e.String()
}
