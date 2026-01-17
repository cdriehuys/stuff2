package i18n

import "context"

type contextKey string

const contextKeyTranslator contextKey = "translator"

func AddToContext(ctx context.Context, t Translator) context.Context {
	return context.WithValue(ctx, contextKeyTranslator, t)
}

func FromContext(ctx context.Context) Translator {
	return ctx.Value(contextKeyTranslator).(Translator)
}
