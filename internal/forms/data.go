package forms

import "github.com/cdriehuys/stuff2/internal/validation"

type Form struct {
	Errors validation.Errors
	Fields map[string]Field
}

type Field struct {
	Name   string
	Value  string
	Errors validation.Errors
}

type Error struct {
	Code    string
	Param   string
	Message string
}

func Make() Form {
	return Form{Fields: make(map[string]Field)}
}
