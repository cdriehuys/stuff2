package validation

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
