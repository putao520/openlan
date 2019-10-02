package libol

import "fmt"

type Err struct {
	Code int
	Message string
}

func Errer(message string, v ...interface{}) (this *Err) {
	this = &Err {
		Message: fmt.Sprintf(message, v...),
	}
	return
}

func (this *Err) String() string {
	return fmt.Sprint("%d: %s", this.Code, this.Message)
}

func (this *Err) Error() string {
	return this.String()
}



