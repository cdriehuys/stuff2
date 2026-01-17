package i18n

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
)

func LoadTranslations(logger *slog.Logger, translations fs.FS) (*ut.UniversalTranslator, error) {
	en := en.New()
	utrans := ut.New(en, en)

	walker := func(path string, d fs.DirEntry, err error) (retErr error) {
		// If there's an error reading a specific path, skip that path.
		if err != nil {
			return nil
		}

		if !d.Type().IsRegular() {
			logger.Debug("Path is not a file with translations.", "path", path)
			return nil
		}

		if filepath.Ext(path) != ".json" {
			logger.Debug("Path is not a JSON file with translations.", "path", path)
			return nil
		}

		f, err := translations.Open(path)
		if err != nil {
			return fmt.Errorf("opening %q: %v", path, err)
		}

		defer func() {
			if err := f.Close(); err != nil {
				retErr = errors.Join(retErr, fmt.Errorf("closing %q: %v", path, err))
			}
		}()

		if err := utrans.ImportByReader(ut.FormatJSON, f); err != nil {
			return fmt.Errorf("loading translations from %q: %v", path, err)
		}

		logger.Info("Loaded translations.", "path", path)

		return nil
	}

	if err := fs.WalkDir(translations, ".", walker); err != nil {
		return nil, fmt.Errorf("finding translations: %v", err)
	}

	if err := utrans.VerifyTranslations(); err != nil {
		return nil, fmt.Errorf("verifying translations: %v", err)
	}

	return utrans, nil
}
