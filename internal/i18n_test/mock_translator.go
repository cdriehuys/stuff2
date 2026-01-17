package i18n_test

import (
	"context"
	"fmt"
	"time"

	"github.com/cdriehuys/stuff2/internal/i18n"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/currency"
	ut "github.com/go-playground/universal-translator"
)

type MockTranslator struct {
	base ut.Translator
}

// CardinalPluralRule implements [i18n.Translator].
func (t MockTranslator) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	panic("unimplemented")
}

// Currency implements [i18n.Translator].
func (t MockTranslator) Currency() currency.Type {
	panic("unimplemented")
}

// FmtAccounting implements [i18n.Translator].
func (t MockTranslator) FmtAccounting(num float64, v uint64, currency currency.Type) string {
	panic("unimplemented")
}

// FmtCurrency implements [i18n.Translator].
func (t MockTranslator) FmtCurrency(num float64, v uint64, currency currency.Type) string {
	panic("unimplemented")
}

// FmtDateFull implements [i18n.Translator].
func (MockTranslator) FmtDateFull(t time.Time) string {
	panic("unimplemented")
}

// FmtDateLong implements [i18n.Translator].
func (MockTranslator) FmtDateLong(t time.Time) string {
	panic("unimplemented")
}

// FmtDateMedium implements [i18n.Translator].
func (MockTranslator) FmtDateMedium(t time.Time) string {
	panic("unimplemented")
}

// FmtDateShort implements [i18n.Translator].
func (MockTranslator) FmtDateShort(t time.Time) string {
	panic("unimplemented")
}

// FmtPercent implements [i18n.Translator].
func (t MockTranslator) FmtPercent(num float64, v uint64) string {
	panic("unimplemented")
}

// FmtTimeFull implements [i18n.Translator].
func (MockTranslator) FmtTimeFull(t time.Time) string {
	panic("unimplemented")
}

// FmtTimeLong implements [i18n.Translator].
func (MockTranslator) FmtTimeLong(t time.Time) string {
	panic("unimplemented")
}

// FmtTimeMedium implements [i18n.Translator].
func (MockTranslator) FmtTimeMedium(t time.Time) string {
	panic("unimplemented")
}

// FmtTimeShort implements [i18n.Translator].
func (MockTranslator) FmtTimeShort(t time.Time) string {
	panic("unimplemented")
}

// Locale implements [i18n.Translator].
func (t MockTranslator) Locale() string {
	panic("unimplemented")
}

// MonthAbbreviated implements [i18n.Translator].
func (t MockTranslator) MonthAbbreviated(month time.Month) string {
	panic("unimplemented")
}

// MonthNarrow implements [i18n.Translator].
func (t MockTranslator) MonthNarrow(month time.Month) string {
	panic("unimplemented")
}

// MonthWide implements [i18n.Translator].
func (t MockTranslator) MonthWide(month time.Month) string {
	panic("unimplemented")
}

// MonthsAbbreviated implements [i18n.Translator].
func (t MockTranslator) MonthsAbbreviated() []string {
	panic("unimplemented")
}

// MonthsNarrow implements [i18n.Translator].
func (t MockTranslator) MonthsNarrow() []string {
	panic("unimplemented")
}

// MonthsWide implements [i18n.Translator].
func (t MockTranslator) MonthsWide() []string {
	panic("unimplemented")
}

// OrdinalPluralRule implements [i18n.Translator].
func (t MockTranslator) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	panic("unimplemented")
}

// PluralsCardinal implements [i18n.Translator].
func (t MockTranslator) PluralsCardinal() []locales.PluralRule {
	panic("unimplemented")
}

// PluralsOrdinal implements [i18n.Translator].
func (t MockTranslator) PluralsOrdinal() []locales.PluralRule {
	panic("unimplemented")
}

// PluralsRange implements [i18n.Translator].
func (t MockTranslator) PluralsRange() []locales.PluralRule {
	panic("unimplemented")
}

// RangePluralRule implements [i18n.Translator].
func (t MockTranslator) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	panic("unimplemented")
}

// WeekdayAbbreviated implements [i18n.Translator].
func (t MockTranslator) WeekdayAbbreviated(weekday time.Weekday) string {
	panic("unimplemented")
}

// WeekdayNarrow implements [i18n.Translator].
func (t MockTranslator) WeekdayNarrow(weekday time.Weekday) string {
	panic("unimplemented")
}

// WeekdayShort implements [i18n.Translator].
func (t MockTranslator) WeekdayShort(weekday time.Weekday) string {
	panic("unimplemented")
}

// WeekdayWide implements [i18n.Translator].
func (t MockTranslator) WeekdayWide(weekday time.Weekday) string {
	panic("unimplemented")
}

// WeekdaysAbbreviated implements [i18n.Translator].
func (t MockTranslator) WeekdaysAbbreviated() []string {
	panic("unimplemented")
}

// WeekdaysNarrow implements [i18n.Translator].
func (t MockTranslator) WeekdaysNarrow() []string {
	panic("unimplemented")
}

// WeekdaysShort implements [i18n.Translator].
func (t MockTranslator) WeekdaysShort() []string {
	panic("unimplemented")
}

// WeekdaysWide implements [i18n.Translator].
func (t MockTranslator) WeekdaysWide() []string {
	panic("unimplemented")
}

func (t MockTranslator) T(key any, params ...string) string {
	return fmt.Sprintf("%v", key)
}

func (t MockTranslator) C(key any, num float64, digits uint64, param string) string {
	return fmt.Sprintf("%v", key)
}

func (t MockTranslator) O(key any, num float64, digits uint64, param string) string {
	return fmt.Sprintf("%v", key)
}

func (t MockTranslator) R(key any, num1 float64, digits1 uint64, num2 float64, digits2 uint64, param1 string, param2 string) string {
	return fmt.Sprintf("%v", key)
}

func (t MockTranslator) Base() ut.Translator {
	return t.base
}

func (t MockTranslator) FmtNumber(num float64, v uint64) string {
	return fmt.Sprintf("%f", num)
}

func WithMockTranslator(ctx context.Context) context.Context {
	return i18n.AddToContext(ctx, MockTranslator{})
}
