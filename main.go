package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cdriehuys/stuff2/internal/application"
	"github.com/cdriehuys/stuff2/internal/email"
	"github.com/cdriehuys/stuff2/internal/i18n"
	"github.com/cdriehuys/stuff2/internal/models"
	"github.com/cdriehuys/stuff2/internal/models/queries"
	"github.com/cdriehuys/stuff2/internal/security"
	"github.com/cdriehuys/stuff2/internal/templating"
	"github.com/cdriehuys/stuff2/translations"
	"github.com/cdriehuys/stuff2/ui"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	emailVerificationTokenLifetime time.Duration = 15 * time.Minute
)

var (
	liveEmailTemplatePath string
	liveTemplatePath      string
)

func main() {
	flag.StringVar(&liveEmailTemplatePath, "live-email-templates", "", "load email templates from this path for each request instead of using the embedded templates")
	flag.StringVar(&liveTemplatePath, "live-templates", "", "load UI templates from this path for each request instead of using the embedded templates")
	flag.Parse()

	logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	var emailTemplates application.TemplateEngine
	if liveEmailTemplatePath != "" {
		emailTemplates = &templating.LiveEmailLoader{Logger: logger, BaseDir: liveEmailTemplatePath}
	} else {
		templateFS, err := fs.Sub(ui.EmailFS, "emails")
		if err != nil {
			panic(err)
		}

		emailTemplates, err = templating.NewEmailTemplateCache(logger, templateFS)
		if err != nil {
			panic(err)
		}
	}

	var uiTemplates application.TemplateEngine
	if liveTemplatePath != "" {
		uiTemplates = &templating.LiveLoader{Logger: logger, BaseDir: liveTemplatePath}
	} else {
		templateFS, err := fs.Sub(ui.FS, "templates")
		if err != nil {
			panic(err)
		}

		uiTemplates, err = templating.NewTemplateCache(logger, templateFS)
		if err != nil {
			panic(err)
		}
	}

	emailer := email.NewConsoleMailer(os.Stdout)
	sender := "no-reply@localhost"
	baseDomain, err := url.Parse("http://localhost:8080")
	if err != nil {
		panic(err)
	}

	emailVerifier := application.NewEmailVerifier(logger, emailer, emailTemplates, baseDomain, sender)

	ut, err := i18n.LoadTranslations(logger, translations.FS)
	if err != nil {
		panic(err)
	}

	connString := os.Getenv("DB_CONN")
	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		panic(err)
	}

	defer dbPool.Close()

	queries := queries.New(dbPool)

	users := models.NewUserModel(
		logger,
		emailVerifier,
		security.Argon2IDHasher{},
		security.TokenGenerator{},
		emailVerificationTokenLifetime,
		models.PoolWrapper{Pool: dbPool},
		models.UserQueriesWrapper{Queries: queries},
	)

	app := application.Application{
		Logger:     logger,
		Templates:  uiTemplates,
		Translator: ut,

		Users: users,
	}

	s := http.Server{
		Addr:    ":8080",
		Handler: app.Routes(),
	}

	logger.Info("Starting web server.", "addr", s.Addr)

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Server stopped.", "error", err)
	}
}
