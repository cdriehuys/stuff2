package validation

type Errors struct {
	errs []Error
}

func NewErrors(errors ...Error) *Errors {
	return &Errors{errs: errors}
}

func (e *Errors) AddNew(code string, message string) {
	e.errs = append(e.errs, Error{
		code:    code,
		message: message,
	})
}

func (e *Errors) Errors() []Error {
	return e.errs
}

func (e *Errors) HasError() bool {
	return len(e.errs) > 0
}

type Error struct {
	code    string
	message string
}

func MakeError(code string, message string) Error {
	return Error{code, message}
}

func (e Error) Code() string {
	return e.code
}

func (e Error) Message() string {
	return e.message
}
