package i18n

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
	ut "github.com/go-playground/universal-translator"
)

type Translator interface {
	locales.Translator

	T(key any, params ...string) string

	C(key any, num float64, digits uint64, param string) string

	O(key any, num float64, digits uint64, param string) string

	R(key any, num1 float64, digits1 uint64, num2 float64, digits2 uint64, param1 string, param2 string) string

	Currency() currency.Type

	Base() ut.Translator
}

type RequestTranslator struct {
	logger *slog.Logger

	locales.Translator
	trans ut.Translator
}

func NewRequestTranslator(logger *slog.Logger, utrans *ut.UniversalTranslator, r *http.Request) *RequestTranslator {
	// If we actually supported multiple languages, this is where we would determine the locale
	// based on request parameters or headers. We don't, so this is easy.

	t, _ := utrans.FindTranslator("en")

	return &RequestTranslator{
		logger:     logger,
		Translator: t.(locales.Translator),
		trans:      t,
	}
}

var _ Translator = (*RequestTranslator)(nil)

func (t *RequestTranslator) T(key any, params ...string) string {
	msg, err := t.trans.T(key, params...)
	if err != nil {
		t.logger.Error("Couldn't translate plain message.", "key", key, "params", params, "error", err)
		return fmt.Sprintf("%v", key)
	}

	return msg
}

func (t *RequestTranslator) C(key any, num float64, digits uint64, param string) string {
	msg, err := t.trans.C(key, num, digits, param)
	if err != nil {
		t.logger.Error("Couldn't translate cardinal message.", "key", key, "num", num, "digits", digits, "param", param, "error", err)
		return fmt.Sprintf("%v", key)
	}

	return msg
}

func (t *RequestTranslator) O(key any, num float64, digits uint64, param string) string {
	msg, err := t.trans.O(key, num, digits, param)
	if err != nil {
		t.logger.Error("Couldn't translate ordinal message.", "key", key, "num", num, "digits", digits, "param", param, "error", err)
		return fmt.Sprintf("%v", key)
	}

	return msg
}

func (t *RequestTranslator) R(key any, num1 float64, digits1 uint64, num2 float64, digits2 uint64, param1 string, param2 string) string {
	msg, err := t.trans.R(key, num1, digits1, num2, digits2, param1, param2)
	if err != nil {
		t.logger.Error(
			"Couldn't translate range message.",
			"key",
			key,
			"num1",
			num1,
			"digits1",
			digits1,
			"num2",
			num2,
			"digits2",
			digits2,
			"param1",
			param1,
			"param2",
			param2,
			"error",
			err,
		)
		return fmt.Sprintf("%v", key)
	}

	return msg
}

func (t *RequestTranslator) Currency() currency.Type {
	return currency.USD
}

func (t *RequestTranslator) Base() ut.Translator {
	return t.trans
}
