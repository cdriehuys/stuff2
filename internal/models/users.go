package models

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cdriehuys/stuff2/internal/i18n"
	"github.com/cdriehuys/stuff2/internal/models/queries"
	"github.com/cdriehuys/stuff2/internal/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NewUser struct {
	Email    string
	Password string
}

type NewUserErrors struct {
	Email    []validation.Error
	Password []validation.Error
}

func (e NewUserErrors) Error() string {
	return fmt.Sprintf("%#v", e)
}

func MakeNewUser(ctx context.Context, email string, password string) (NewUser, error) {
	t := i18n.FromContext(ctx)

	validationErrors := NewUserErrors{}

	trimmedEmail := strings.TrimSpace(email)
	if len(trimmedEmail) == 0 {
		validationErrors.Email = append(validationErrors.Email, validation.MakeError("required", t.T("user.email.required")))
	} else if len(trimmedEmail) < 3 || len(trimmedEmail) > 254 || !strings.Contains(trimmedEmail, "@") {
		validationErrors.Email = append(validationErrors.Email, validation.MakeError("email", t.T("user.email.invalid")))
	}

	if len(password) < 8 {
		validationErrors.Password = append(validationErrors.Password, validation.MakeError("min", t.C("user.password.length.min", 8, 0, t.FmtNumber(8, 0))))
	} else if len(password) > 1000 {
		validationErrors.Password = append(validationErrors.Password, validation.MakeError("max", t.C("user.password.length.max", 1000, 0, t.FmtNumber(1000, 0))))
	}

	if len(validationErrors.Email) > 0 || len(validationErrors.Password) > 0 {
		return NewUser{}, validationErrors
	}

	return NewUser{trimmedEmail, password}, nil
}

type User struct {
	ID uuid.UUID
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	ComparePasswordAndHash(password string, hash string) (bool, error)
}

type TokenGenerator interface {
	Generate() string
}

type EmailVerifier interface {
	DuplicateRegistration(ctx context.Context, email string) error
	NewEmail(ctx context.Context, email string, token string) error
}

type UserQueries interface {
	WithTx(tx queries.DBTX) UserQueries

	DeleteEmailVerificationKeyByID(ctx context.Context, id int32) error
	DeleteUnverifiedEmails(ctx context.Context, email string) error
	GetEmailVerificationKeyByToken(context.Context, string) (queries.EmailVerificationKey, error)
	GetUserByVerifiedEmail(ctx context.Context, email string) (queries.User, error)
	InsertEmailVerificationKey(context.Context, queries.InsertEmailVerificationKeyParams) error
	InsertNewUser(context.Context, queries.InsertNewUserParams) (queries.User, error)
	VerifiedEmailExists(context.Context, string) (bool, error)
	VerifyEmailForUser(ctx context.Context, userID uuid.UUID) error
}

type UserQueriesWrapper struct {
	*queries.Queries
}

func (w UserQueriesWrapper) WithTx(tx queries.DBTX) UserQueries {
	return UserQueriesWrapper{w.Queries.WithTx(tx.(pgx.Tx))}
}

type UserModel struct {
	logger         *slog.Logger
	emailVerifier  EmailVerifier
	hasher         PasswordHasher
	tokenGenerator TokenGenerator
	tokenLifetime  time.Duration

	db DB
	q  UserQueries
}

func NewUserModel(
	logger *slog.Logger,
	emailVerifier EmailVerifier,
	hasher PasswordHasher,
	tokenGenerator TokenGenerator,
	tokenLifetime time.Duration,
	db DB,
	queries UserQueries,
) *UserModel {
	return &UserModel{
		logger:         logger,
		emailVerifier:  emailVerifier,
		hasher:         hasher,
		tokenGenerator: tokenGenerator,
		tokenLifetime:  tokenLifetime,
		db:             db,
		q:              queries,
	}
}

var ErrInvalidCredentials = errors.New("invalid credentials")

const dummyComparePassword = "jekyll"

// dummyCompareHash is the result of running argon2id.CreateHash("hyde", argon2id.DefaultParams)
const dummyCompareHash = "$argon2id$v=19$m=65536,t=1,p=8$+6TY8oCAV6WfG6DPgK55Lg$paDhSsjLEhkbUb8YOha73/zxzuC9VJgxCGvKosEtOEQ"

func (m *UserModel) Authenticate(ctx context.Context, email string, password string) (User, error) {
	user, err := m.q.GetUserByVerifiedEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Do a password/hash comparison to mitigate timing attacks. The values don't matter as
			// long as the comparison hash can be decoded as a hash.
			m.hasher.ComparePasswordAndHash(dummyComparePassword, dummyCompareHash)

			return User{}, ErrInvalidCredentials
		}

		return User{}, fmt.Errorf("searching for user: %v", err)
	}

	passwordMatches, err := m.hasher.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		// An error would be returned if a hash is somehow malformed. This is different from a
		// password that does not match the given hash.
		return User{}, fmt.Errorf("comparing password to hash: %v", err)
	}

	if !passwordMatches {
		return User{}, ErrInvalidCredentials
	}

	return User{ID: user.ID}, nil
}

func (m *UserModel) Register(ctx context.Context, user NewUser) (retErr error) {
	passwordHash, err := m.hasher.Hash(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %v", err)
	}

	defer func() {
		if txErr := tx.Rollback(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
			retErr = errors.Join(retErr, txErr)
		}
	}()

	txQueries := m.q.WithTx(tx)

	emailAlreadyVerified, err := txQueries.VerifiedEmailExists(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("failed to check for duplicate email: %v", err)
	}

	if emailAlreadyVerified {
		m.logger.DebugContext(ctx, "Registration is for an email that has already been verified.")

		// Let the transaction roll back because there are no changes.
		return m.emailVerifier.DuplicateRegistration(ctx, user.Email)
	}

	userID := uuid.New()

	userParams := queries.InsertNewUserParams{
		ID:           userID,
		Email:        user.Email,
		PasswordHash: passwordHash,
	}
	if _, err := txQueries.InsertNewUser(ctx, userParams); err != nil {
		return fmt.Errorf("failed to persist new user: %v", err)
	}

	m.logger.DebugContext(ctx, "Persisted new user.", "userID", userID)

	verificationToken := m.tokenGenerator.Generate()

	emailVerificationParams := queries.InsertEmailVerificationKeyParams{
		UserID: userID,
		Email:  user.Email,
		Token:  verificationToken,
	}
	if err := txQueries.InsertEmailVerificationKey(ctx, emailVerificationParams); err != nil {
		return fmt.Errorf("failed to insert email verification key: %v", err)
	}

	m.logger.DebugContext(ctx, "Persisted email verification key.", "userID", userID)

	if err := m.emailVerifier.NewEmail(ctx, user.Email, verificationToken); err != nil {
		return fmt.Errorf("failed to send email verification: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit user registration: %v", err)
	}

	m.logger.InfoContext(ctx, "Registered a new user.", "userID", userID)

	return nil
}

var ErrInvalidEmailVerificationToken = errors.New("invalid token")

func (m *UserModel) VerifyEmail(ctx context.Context, token string) (retErr error) {
	// 1. Get token
	// 2. Check expiration
	// [in tx]
	// 3. Mark verified
	// 4. Delete others
	// [end tx]
	// 5. Delete verification
	verification, err := m.q.GetEmailVerificationKeyByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			m.logger.DebugContext(ctx, "Email verification token does not exist.")

			return ErrInvalidEmailVerificationToken
		}

		return fmt.Errorf("retrieving verification key: %v", err)
	}

	if verification.CreatedAt.Time.Add(m.tokenLifetime).Before(time.Now()) {
		m.logger.DebugContext(ctx, "Email verification token is expired.")

		return ErrInvalidEmailVerificationToken
	}

	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %v", err)
	}

	defer func() {
		if txErr := tx.Rollback(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
			retErr = errors.Join(retErr, txErr)
		}
	}()

	txQueries := m.q.WithTx(tx)

	if err := txQueries.VerifyEmailForUser(ctx, verification.UserID); err != nil {
		return fmt.Errorf("marking email verified for user %s: %v", verification.UserID.String(), err)
	}

	if err := txQueries.DeleteUnverifiedEmails(ctx, verification.Email); err != nil {
		return fmt.Errorf("deleting duplicate unverified users: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %v", err)
	}

	m.logger.InfoContext(ctx, "Verified email address for user.", "userID", verification.UserID)

	if err := m.q.DeleteEmailVerificationKeyByID(ctx, verification.ID); err != nil {
		return fmt.Errorf("deleting used verification key: %v", err)
	}

	return nil
}
