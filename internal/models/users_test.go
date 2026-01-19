package models_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/cdriehuys/stuff2/internal/i18n_test"
	"github.com/cdriehuys/stuff2/internal/models"
	"github.com/cdriehuys/stuff2/internal/models/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestMakeNewUser(t *testing.T) {
	type wantCodes struct {
		email    []string
		password []string
	}

	testCases := []struct {
		name         string
		email        string
		password     string
		wantSuccess  bool
		wantEmail    string
		wantPassword string
		wantCodes    wantCodes
	}{
		{
			name:     "empty",
			email:    "",
			password: "",
			wantCodes: wantCodes{
				email:    []string{"required"},
				password: []string{"min"},
			},
		},
		{
			name:     "email missing @",
			email:    "localhost",
			password: "validpassword",
			wantCodes: wantCodes{
				email: []string{"email"},
			},
		},
		{
			name:     "email too short",
			email:    "ab",
			password: "validpassword",
			wantCodes: wantCodes{
				email: []string{"email"},
			},
		},
		{
			name:     "email too long",
			email:    "abc@" + strings.Repeat("d", 251),
			password: "validpassword",
			wantCodes: wantCodes{
				email: []string{"email"},
			},
		},
		{
			name:     "password too short",
			email:    "test@example.com",
			password: strings.Repeat("a", 7),
			wantCodes: wantCodes{
				password: []string{"min"},
			},
		},
		{
			name:     "password too long",
			email:    "test@example.com",
			password: strings.Repeat("a", 1001),
			wantCodes: wantCodes{
				password: []string{"max"},
			},
		},
		{
			name:         "valid data",
			email:        "test@example.com",
			password:     "password",
			wantSuccess:  true,
			wantEmail:    "test@example.com",
			wantPassword: "password",
		},
		{
			name:        "valid trimmed email",
			email:       " test@example.com ",
			password:    " password ",
			wantSuccess: true,
			wantEmail:   "test@example.com",
			// Password should not be trimmed.
			wantPassword: " password ",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := i18n_test.WithMockTranslator(t.Context())

			user, err := models.MakeNewUser(ctx, tt.email, tt.password)

			if tt.wantEmail != user.Email {
				t.Errorf("Expected user email %q, got %q", tt.wantEmail, user.Email)
			}

			if tt.wantPassword != user.Password {
				t.Errorf("Expected user password %q, got %q", tt.wantPassword, user.Password)
			}

			if tt.wantSuccess {
				if err != nil {
					t.Errorf("Expected success, got error %v", err)
				}

				return
			}

			userErrs := models.NewUserErrors{}
			if !errors.As(err, &userErrs) {
				t.Fatalf("Expected `NewUserErrors{}`, got %#v", err)
			}

			if len(tt.wantCodes.email) != len(userErrs.Email) {
				t.Errorf("Expected error codes %v, got %v", tt.wantCodes.email, userErrs.Email)
			}

			gotEmailCodes := make(map[string]struct{}, len(tt.wantCodes.email))
			for _, emailErr := range userErrs.Email {
				gotEmailCodes[emailErr.Code()] = struct{}{}
			}

			for _, wantCode := range tt.wantCodes.email {
				if _, exists := gotEmailCodes[wantCode]; !exists {
					t.Errorf("Expected code %q in %v", wantCode, gotEmailCodes)
				}
			}

			if len(tt.wantCodes.password) != len(userErrs.Password) {
				t.Errorf("Expected password error codes %v, got %v", tt.wantCodes.password, userErrs.Password)
			}

			gotPasswordCodes := make(map[string]struct{}, len(tt.wantCodes.password))
			for _, passwordErr := range userErrs.Password {
				gotPasswordCodes[passwordErr.Code()] = struct{}{}
			}

			for _, wantCode := range tt.wantCodes.password {
				if _, exists := gotPasswordCodes[wantCode]; !exists {
					t.Errorf("Expected code %q in %v", wantCode, gotPasswordCodes)
				}
			}
		})
	}
}

const mockHashValue = "hashed"
const mockToken = "secret-token"

var defaultNewUser = models.NewUser{
	Email:    "test@example.com",
	Password: "tops3cret",
}

var errInsert = errors.New("insert failed")
var errRollback = errors.New("rollback failed")

type ConstantHasher struct {
	HashError    error
	CompareError error
}

func (h ConstantHasher) Hash(string) (string, error) {
	return mockHashValue, h.HashError
}

func (h ConstantHasher) ComparePasswordAndHash(password string, hash string) (bool, error) {
	return password == hash, h.CompareError
}

type ConstantTokenGenerator struct {
	token string
}

func (g *ConstantTokenGenerator) Generate() string {
	return g.token
}

type MockEmailVerifier struct {
	duplicateRegistrationEmail string
	duplicateRegistrationError error

	newEmailEmail string
	newEmailToken string
	newEmailError error
}

func (v *MockEmailVerifier) DuplicateRegistration(ctx context.Context, email string) error {
	v.duplicateRegistrationEmail = email
	return v.duplicateRegistrationError
}

func (v *MockEmailVerifier) NewEmail(ctx context.Context, email string, token string) error {
	v.newEmailEmail = email
	v.newEmailToken = token

	return v.newEmailError
}

type MockUserQueries struct {
	deletedEmailVerificationID          int32
	deleteEmailVerificationKeyByIDError error

	deleteUnverifiedEmailsEmail string
	deleteUnverifiedEmailsError error

	getEmailVerificationKeyByTokenToken  string
	getEmailVerificationKeyByTokenReturn queries.EmailVerificationKey
	GetEmailVerificationKeyByTokenError  error

	insertEmailVerificationKeyError error
	insertEmailVerificationParams   queries.InsertEmailVerificationKeyParams

	insertNewUserReturnUser  queries.User
	insertNewUserReturnError error
	insertNewUserParams      queries.InsertNewUserParams

	verifiedEmailExistsEmail  string
	verifiedEmailExistsReturn bool
	verifiedEmailExistsError  error

	verifyEmailForUserID    uuid.UUID
	verifyEmailForUserError error
}

func (q *MockUserQueries) WithTx(queries.DBTX) models.UserQueries {
	return q
}

func (q *MockUserQueries) DeleteEmailVerificationKeyByID(ctx context.Context, id int32) error {
	q.deletedEmailVerificationID = id

	return q.deleteEmailVerificationKeyByIDError
}

func (q *MockUserQueries) DeleteUnverifiedEmails(ctx context.Context, email string) error {
	q.deleteUnverifiedEmailsEmail = email

	return q.deleteUnverifiedEmailsError
}

func (q *MockUserQueries) GetEmailVerificationKeyByToken(ctx context.Context, token string) (queries.EmailVerificationKey, error) {
	q.getEmailVerificationKeyByTokenToken = token

	return q.getEmailVerificationKeyByTokenReturn, q.GetEmailVerificationKeyByTokenError
}

func (q *MockUserQueries) InsertEmailVerificationKey(ctx context.Context, params queries.InsertEmailVerificationKeyParams) error {
	q.insertEmailVerificationParams = params

	return q.insertEmailVerificationKeyError
}

func (q *MockUserQueries) InsertNewUser(ctx context.Context, params queries.InsertNewUserParams) (queries.User, error) {
	q.insertNewUserParams = params

	return q.insertNewUserReturnUser, q.insertNewUserReturnError
}

func (q *MockUserQueries) VerifiedEmailExists(ctx context.Context, email string) (bool, error) {
	q.verifiedEmailExistsEmail = email

	return q.verifiedEmailExistsReturn, q.verifiedEmailExistsError
}

func (q *MockUserQueries) VerifyEmailForUser(ctx context.Context, userID uuid.UUID) error {
	q.verifyEmailForUserID = userID

	return q.verifyEmailForUserError
}

func TestUserModel_Register(t *testing.T) {
	testCases := []struct {
		name                             string
		emailVerifier                    MockEmailVerifier
		hasher                           ConstantHasher
		tokenGenerator                   ConstantTokenGenerator
		db                               MockDB
		tx                               MockTX
		queries                          MockUserQueries
		newUser                          models.NewUser
		wantVerifiedEmailCheck           string
		wantInsertedUser                 queries.InsertNewUserParams
		wantInsertedEmailVerificationKey queries.InsertEmailVerificationKeyParams
		wantNewEmailNotification         string
		wantNewEmailToken                string
		wantDuplicateEmailNotification   string
		wantTxRollback                   bool
		wantTxCommit                     bool
		wantErr                          bool
		wantErrors                       []error
	}{
		{
			name: "hash error",
			hasher: ConstantHasher{
				HashError: errors.New("something happened"),
			},
			newUser: defaultNewUser,
			wantErr: true,
		},
		{
			name: "error starting transaction",
			db: MockDB{
				beginError: errors.New("failed to start tx"),
			},
			newUser: defaultNewUser,
			wantErr: true,
		},
		{
			name:    "error checking for existing email",
			newUser: defaultNewUser,
			queries: MockUserQueries{
				verifiedEmailExistsError: errors.New("query failed"),
			},
			wantVerifiedEmailCheck: defaultNewUser.Email,
			wantTxRollback:         true,
			wantErr:                true,
		},
		{
			name:    "failed user insert",
			newUser: defaultNewUser,
			queries: MockUserQueries{
				insertNewUserReturnError: errors.New("insert failed"),
			},
			wantVerifiedEmailCheck: defaultNewUser.Email,
			wantInsertedUser: queries.InsertNewUserParams{
				Email:        defaultNewUser.Email,
				PasswordHash: mockHashValue,
			},
			wantTxRollback: true,
			wantErr:        true,
		},
		{
			name:    "failed rollback",
			newUser: defaultNewUser,
			tx:      MockTX{rollbackError: errRollback},
			queries: MockUserQueries{
				verifiedEmailExistsError: errInsert,
			},
			wantVerifiedEmailCheck: defaultNewUser.Email,
			wantErr:                true,
			wantErrors:             []error{errRollback, errInsert},
		},
		{
			name: "failed verification token insert",
			tokenGenerator: ConstantTokenGenerator{
				token: mockToken,
			},
			queries: MockUserQueries{
				insertEmailVerificationKeyError: errors.New("insert failed"),
			},
			newUser:                defaultNewUser,
			wantVerifiedEmailCheck: defaultNewUser.Email,
			wantInsertedUser: queries.InsertNewUserParams{
				Email:        defaultNewUser.Email,
				PasswordHash: mockHashValue,
			},
			wantInsertedEmailVerificationKey: queries.InsertEmailVerificationKeyParams{
				Email: defaultNewUser.Email,
				Token: mockToken,
			},
			wantTxRollback: true,
			wantErr:        true,
		},
		{
			name: "new user notification fail",
			emailVerifier: MockEmailVerifier{
				newEmailError: errors.New("failed to send notification"),
			},
			tokenGenerator: ConstantTokenGenerator{
				token: mockToken,
			},
			newUser:                defaultNewUser,
			wantVerifiedEmailCheck: defaultNewUser.Email,
			wantInsertedUser: queries.InsertNewUserParams{
				Email:        defaultNewUser.Email,
				PasswordHash: mockHashValue,
			},
			wantInsertedEmailVerificationKey: queries.InsertEmailVerificationKeyParams{
				Email: defaultNewUser.Email,
				Token: mockToken,
			},
			wantNewEmailNotification: defaultNewUser.Email,
			wantNewEmailToken:        mockToken,
			wantTxRollback:           true,
			wantErr:                  true,
		},
		{
			name: "commit fail",
			tokenGenerator: ConstantTokenGenerator{
				token: mockToken,
			},
			tx: MockTX{
				commitError: errors.New("failed to commit"),
			},
			newUser:                defaultNewUser,
			wantVerifiedEmailCheck: defaultNewUser.Email,
			wantInsertedUser: queries.InsertNewUserParams{
				Email:        defaultNewUser.Email,
				PasswordHash: mockHashValue,
			},
			wantInsertedEmailVerificationKey: queries.InsertEmailVerificationKeyParams{
				Email: defaultNewUser.Email,
				Token: mockToken,
			},
			wantNewEmailNotification: defaultNewUser.Email,
			wantNewEmailToken:        mockToken,
			wantTxRollback:           true,
			wantErr:                  true,
		},
		{
			name: "duplicate user registration",
			queries: MockUserQueries{
				verifiedEmailExistsReturn: true,
			},
			newUser:                        defaultNewUser,
			wantVerifiedEmailCheck:         defaultNewUser.Email,
			wantDuplicateEmailNotification: defaultNewUser.Email,
			wantTxRollback:                 true,
		},
		{
			name: "duplicate user notification error",
			emailVerifier: MockEmailVerifier{
				duplicateRegistrationError: errors.New("notification failed"),
			},
			queries: MockUserQueries{
				verifiedEmailExistsReturn: true,
			},
			newUser:                        defaultNewUser,
			wantVerifiedEmailCheck:         defaultNewUser.Email,
			wantDuplicateEmailNotification: defaultNewUser.Email,
			wantTxRollback:                 true,
			wantErr:                        true,
		},
		{
			name: "new user registration",
			tokenGenerator: ConstantTokenGenerator{
				token: mockToken,
			},
			newUser:                defaultNewUser,
			wantVerifiedEmailCheck: defaultNewUser.Email,
			wantInsertedUser: queries.InsertNewUserParams{
				Email:        defaultNewUser.Email,
				PasswordHash: mockHashValue,
			},
			wantInsertedEmailVerificationKey: queries.InsertEmailVerificationKeyParams{
				Email: defaultNewUser.Email,
				Token: mockToken,
			},
			wantNewEmailNotification: defaultNewUser.Email,
			wantNewEmailToken:        mockToken,
			wantTxCommit:             true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.db.txFactory == nil {
				tt.db.txFactory = func() models.Transaction { return &tt.tx }
			}

			users := models.NewUserModel(
				slog.New(slog.DiscardHandler),
				&tt.emailVerifier,
				&tt.hasher,
				&tt.tokenGenerator,
				time.Minute,
				&tt.db,
				&tt.queries,
			)

			err := users.Register(t.Context(), tt.newUser)

			if err == nil && tt.wantErr {
				t.Fatal("Expected Register to error.")
			}

			if err != nil && !tt.wantErr {
				t.Fatalf("Register returned an error: %#v", err)
			}

			if tt.wantErrors != nil {
				for _, wantErr := range tt.wantErrors {
					if !strings.Contains(err.Error(), wantErr.Error()) {
						t.Errorf("Expected error to include %q, got %q", wantErr.Error(), err.Error())
					}
				}
			}

			if tt.wantTxCommit != tt.tx.committed {
				t.Errorf("Expected tx.committed=%v, got %v", tt.wantTxCommit, tt.tx.committed)
			}

			if tt.wantTxRollback != tt.tx.rolledBack {
				t.Errorf("Expected tx.rolledBack=%v, got %v", tt.wantTxRollback, tt.tx.rolledBack)
			}

			if got := tt.queries.verifiedEmailExistsEmail; got != tt.wantVerifiedEmailCheck {
				t.Errorf("Expected check for verified email %q, got %q", tt.wantVerifiedEmailCheck, got)
			}

			if got := tt.queries.insertNewUserParams.Email; got != tt.wantInsertedUser.Email {
				t.Errorf("Expected inserted email to be %q, got %q", tt.wantInsertedUser.Email, got)
			}

			if got := tt.queries.insertNewUserParams.PasswordHash; got != tt.wantInsertedUser.PasswordHash {
				t.Errorf("Expected password hash %q, got %q", tt.wantInsertedUser.PasswordHash, got)
			}

			if got := tt.queries.insertEmailVerificationParams.UserID; (got != uuid.UUID{}) && got != tt.queries.insertNewUserParams.ID {
				t.Errorf("User ID %v for email verification does not match inserted user %v", got, tt.queries.insertNewUserParams.ID)
			}

			if got := tt.queries.insertEmailVerificationParams; got.Email != tt.wantInsertedEmailVerificationKey.Email {
				t.Errorf("Expected verification email %q, got %q", tt.wantInsertedEmailVerificationKey.Email, got.Email)
			}

			if got := tt.queries.insertEmailVerificationParams; got.Token != tt.wantInsertedEmailVerificationKey.Token {
				t.Errorf("Expected verification token %q, got %q", tt.wantInsertedEmailVerificationKey.Token, got.Token)
			}

			if got := tt.emailVerifier.duplicateRegistrationEmail; got != tt.wantDuplicateEmailNotification {
				t.Errorf("Expected duplicate email notification for %q, got %q", tt.wantDuplicateEmailNotification, got)
			}

			if got := tt.emailVerifier.newEmailEmail; got != tt.wantNewEmailNotification {
				t.Errorf("Expected email verification for %q, got %q", tt.wantNewEmailNotification, got)
			}

			if got := tt.emailVerifier.newEmailToken; got != tt.wantNewEmailToken {
				t.Errorf("Expected email verification token %q, got %q", tt.wantNewEmailToken, got)
			}
		})
	}
}

func TestUserModel_VerifyEmail(t *testing.T) {
	genericDBError := errors.New("generic DB error")
	defaultUserID := uuid.New()

	testCases := []struct {
		name                           string
		db                             MockDB
		tx                             MockTX
		queries                        MockUserQueries
		tokenLifetime                  time.Duration
		token                          string
		wantVerifiedUserID             uuid.UUID
		wantUnverifiedEmailsDeletedFor string
		wantDeletedEmailVerificationID int32
		wantTxRollback                 bool
		wantTxCommit                   bool
		wantErr                        bool
		wantErrors                     []error
	}{
		{
			name: "missing token",
			queries: MockUserQueries{
				GetEmailVerificationKeyByTokenError: pgx.ErrNoRows,
			},
			token:      "missing",
			wantErr:    true,
			wantErrors: []error{models.ErrInvalidEmailVerificationToken},
		},
		{
			name: "error retrieving token",
			queries: MockUserQueries{
				GetEmailVerificationKeyByTokenError: genericDBError,
			},
			token:      "causes-error",
			wantErr:    true,
			wantErrors: []error{genericDBError},
		},
		{
			name: "expired token",
			queries: MockUserQueries{
				getEmailVerificationKeyByTokenReturn: queries.EmailVerificationKey{
					CreatedAt: pgtype.Timestamptz{
						Time: time.Now().Add(-2 * time.Minute),
					},
				},
			},
			tokenLifetime: time.Minute,
			token:         "expired",
			wantErr:       true,
			wantErrors:    []error{models.ErrInvalidEmailVerificationToken},
		},
		{
			name: "error starting transaction",
			db:   MockDB{beginError: genericDBError},
			queries: MockUserQueries{
				getEmailVerificationKeyByTokenReturn: queries.EmailVerificationKey{
					CreatedAt: pgtype.Timestamptz{
						Time: time.Now(),
					},
				},
			},
			tokenLifetime: time.Minute,
			wantErr:       true,
		},
		{
			name: "error marking email as verified",
			queries: MockUserQueries{
				getEmailVerificationKeyByTokenReturn: queries.EmailVerificationKey{
					UserID: defaultUserID,
					CreatedAt: pgtype.Timestamptz{
						Time: time.Now(),
					},
				},
				verifyEmailForUserError: genericDBError,
			},
			tokenLifetime:      time.Minute,
			wantVerifiedUserID: defaultUserID,
			wantTxRollback:     true,
			wantErr:            true,
		},
		{
			name: "transaction rollback error preserves existing",
			tx:   MockTX{rollbackError: errRollback},
			queries: MockUserQueries{
				getEmailVerificationKeyByTokenReturn: queries.EmailVerificationKey{
					UserID: defaultUserID,
					CreatedAt: pgtype.Timestamptz{
						Time: time.Now(),
					},
				},
				verifyEmailForUserError: genericDBError,
			},
			tokenLifetime:      time.Minute,
			wantVerifiedUserID: defaultUserID,
			wantErr:            true,
			wantErrors:         []error{errRollback, genericDBError},
		},
		{
			name: "error deleting unverified emails",
			queries: MockUserQueries{
				deleteUnverifiedEmailsError: genericDBError,
				getEmailVerificationKeyByTokenReturn: queries.EmailVerificationKey{
					UserID: defaultUserID,
					Email:  "test@example.com",
					CreatedAt: pgtype.Timestamptz{
						Time: time.Now(),
					},
				},
			},
			tokenLifetime:                  time.Minute,
			wantVerifiedUserID:             defaultUserID,
			wantUnverifiedEmailsDeletedFor: "test@example.com",
			wantTxRollback:                 true,
			wantErr:                        true,
		},
		{
			name: "error committing transaction",
			tx:   MockTX{commitError: genericDBError},
			queries: MockUserQueries{
				getEmailVerificationKeyByTokenReturn: queries.EmailVerificationKey{
					UserID: defaultUserID,
					Email:  "test@example.com",
					CreatedAt: pgtype.Timestamptz{
						Time: time.Now(),
					},
				},
			},
			tokenLifetime:                  time.Minute,
			wantVerifiedUserID:             defaultUserID,
			wantUnverifiedEmailsDeletedFor: "test@example.com",
			wantTxRollback:                 true,
			wantErr:                        true,
		},
		{
			name: "error deleting verification",
			queries: MockUserQueries{
				deleteEmailVerificationKeyByIDError: genericDBError,
				getEmailVerificationKeyByTokenReturn: queries.EmailVerificationKey{
					ID:     3,
					UserID: defaultUserID,
					Email:  "test@example.com",
					CreatedAt: pgtype.Timestamptz{
						Time: time.Now(),
					},
				},
			},
			tokenLifetime:                  time.Minute,
			wantVerifiedUserID:             defaultUserID,
			wantUnverifiedEmailsDeletedFor: "test@example.com",
			wantDeletedEmailVerificationID: 3,
			wantTxCommit:                   true,
			wantErr:                        true,
		},
		{
			name: "success",
			queries: MockUserQueries{
				getEmailVerificationKeyByTokenReturn: queries.EmailVerificationKey{
					ID:     42,
					UserID: defaultUserID,
					Email:  "test@example.com",
					CreatedAt: pgtype.Timestamptz{
						Time: time.Now(),
					},
				},
			},
			tokenLifetime:                  time.Minute,
			wantVerifiedUserID:             defaultUserID,
			wantUnverifiedEmailsDeletedFor: "test@example.com",
			wantDeletedEmailVerificationID: 42,
			wantTxCommit:                   true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.db.txFactory == nil {
				tt.db.txFactory = func() models.Transaction { return &tt.tx }
			}

			users := models.NewUserModel(
				slog.New(slog.DiscardHandler),
				&MockEmailVerifier{},
				ConstantHasher{},
				&ConstantTokenGenerator{},
				tt.tokenLifetime,
				&tt.db,
				&tt.queries,
			)

			err := users.VerifyEmail(t.Context(), tt.token)

			if err == nil && tt.wantErr {
				t.Fatal("Expected VerifyEmail to error.")
			}

			if err != nil && !tt.wantErr {
				t.Fatalf("VerifyEmail returned an error: %#v", err)
			}

			if tt.wantErrors != nil {
				for _, wantErr := range tt.wantErrors {
					if !strings.Contains(err.Error(), wantErr.Error()) {
						t.Errorf("Expected error to include %q, got %q", wantErr.Error(), err.Error())
					}
				}
			}

			if tt.wantTxCommit != tt.tx.committed {
				t.Errorf("Expected tx.committed=%v, got %v", tt.wantTxCommit, tt.tx.committed)
			}

			if tt.wantTxRollback != tt.tx.rolledBack {
				t.Errorf("Expected tx.rolledBack=%v, got %v", tt.wantTxRollback, tt.tx.rolledBack)
			}

			if got := tt.queries.getEmailVerificationKeyByTokenToken; got != tt.token {
				t.Errorf("Expected query for verification token %q, got %q", tt.token, got)
			}

			if got := tt.queries.verifyEmailForUserID; got != tt.wantVerifiedUserID {
				t.Errorf("Expected verified user ID %q, got %q", tt.wantVerifiedUserID, got)
			}

			if got := tt.queries.deleteUnverifiedEmailsEmail; got != tt.wantUnverifiedEmailsDeletedFor {
				t.Errorf("Expected unverified instances of %q to be deleted, got %q", tt.wantUnverifiedEmailsDeletedFor, got)
			}

			if got := tt.queries.deletedEmailVerificationID; got != tt.wantDeletedEmailVerificationID {
				t.Errorf("Expected email verification key %d to be deleted, got %d", tt.wantDeletedEmailVerificationID, got)
			}
		})
	}
}
