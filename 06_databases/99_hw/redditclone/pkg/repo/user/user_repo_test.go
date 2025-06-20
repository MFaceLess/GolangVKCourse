package repo

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"golang.org/x/crypto/bcrypt"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	"redditclone/pkg/user"
)

type fakeJWT struct {
	token string
	err   error
}

func (f *fakeJWT) GenerateJWT(username, userID string) (string, error) {
	return f.token, f.err
}

type failHasher struct{}

func (failHasher) HashPassword(pass string) ([]byte, error) {
	return nil, fmt.Errorf("hash error")
}

type failIDGen struct{}

func (failIDGen) GenerateID() ([]byte, error) {
	return nil, fmt.Errorf("rand fail")
}

func TestAuthorize(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	pass := "secret"
	hp, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("cant generate hash: %s", err)
	}

	rows := sqlmock.NewRows([]string{"id", "password"})
	expectedUser := &user.User{Login: "temp", ID: "id", Password: hp}
	rows.AddRow(expectedUser.ID, hp)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, password FROM users WHERE login = ?")).
		WithArgs("temp").
		WillReturnRows(rows)

	jwtGen := &fakeJWT{token: "token", err: nil}
	repo := NewMemoryRepo(db, jwtGen)

	token, err := repo.Authorize("temp", pass)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if token != "token" {
		t.Errorf("expected token, got %q", token)
	}

	if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
		t.Errorf("there were unfulfilled expectations: %s", mockErr)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, password FROM users WHERE login = ?")).
		WithArgs("not_exist").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Authorize("not_exist", pass)

	if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
		t.Errorf("there were unfulfilled expectations: %s", mockErr)
	}
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if err != ErrNotFoundUser {
		t.Errorf("expected error %v, but got %v", ErrNotFoundUser, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, password FROM users WHERE login = ?")).
		WithArgs("unknowErr").
		WillReturnError(sql.ErrTxDone)

	_, err = repo.Authorize("unknowErr", pass)

	if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
		t.Errorf("there were unfulfilled expectations: %s", mockErr)
	}
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	rows = sqlmock.NewRows([]string{"id", "password"}).AddRow("id", hp)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, password FROM users WHERE login = ?")).
		WithArgs("temp").
		WillReturnRows(rows)

	_, err = repo.Authorize("temp", string([]byte("error")))

	if mockErr := mock.ExpectationsWereMet(); mockErr != nil {
		t.Errorf("there were unfulfilled expectations: %s", mockErr)
	}
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if err != ErrNotFoundUser {
		t.Errorf("expected error %v, but got %v", ErrNotFoundUser, err)
	}
}

func TestRegister(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	jwtGen := &fakeJWT{token: "token", err: nil}
	repo := NewMemoryRepo(db, jwtGen)

	t.Run("successful register", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
			WithArgs("newuser").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (id, login, password) VALUES (?, ?, ?)")).
			WithArgs(sqlmock.AnyArg(), "newuser", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		token, err := repo.Register("newuser", "12345")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if token != "token" {
			t.Errorf("expected token, got %s", token)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
			WithArgs("newuser").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		_, err := repo.Register("newuser", "12345")
		if err != ErrUserAlreadyExists {
			t.Errorf("expected ErrUserAlreadyExists, got %v", err)
		}
	})

	t.Run("query error on exists check", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
			WithArgs("dberror").
			WillReturnError(sql.ErrConnDone)

		_, err := repo.Register("dberror", "password123")
		if err != sql.ErrConnDone {
			t.Errorf("expected sql.ErrConnDone, got %v", err)
		}
	})

	t.Run("insert error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
			WithArgs("insertfail").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (id, login, password) VALUES (?, ?, ?)")).
			WithArgs(sqlmock.AnyArg(), "insertfail", sqlmock.AnyArg()).
			WillReturnError(sql.ErrTxDone)

		_, err := repo.Register("insertfail", "password123")
		if err != sql.ErrTxDone {
			t.Errorf("expected sql.ErrTxDone, got %v", err)
		}
	})

	t.Run("jwt error", func(t *testing.T) {
		errorJWT := &fakeJWT{token: "", err: fmt.Errorf("jwt error")}
		repo := NewMemoryRepo(db, errorJWT)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
			WithArgs("jwtfail").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (id, login, password) VALUES (?, ?, ?)")).
			WithArgs(sqlmock.AnyArg(), "jwtfail", sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := repo.Register("jwtfail", "password123")
		if err == nil || err.Error() != "jwt error" {
			t.Errorf("expected jwt error, got %v", err)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestErrorHashAndErrorGenerateID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	jwtGen := &fakeJWT{token: "token", err: nil}
	repo := &UserMemoryRepository{
		db:      db,
		session: jwtGen,
		hasher:  failHasher{},
		idgen:   failIDGen{},
	}

	t.Run("hash error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
			WithArgs("newuser").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		_, err := repo.Register("newuser", "12345")
		if err == nil || err != ErrCantGenerateHashPassword {
			t.Errorf("expected error: %v, but got: %v", ErrCantGenerateHashPassword, err)
		}
	})

	repo = &UserMemoryRepository{
		db:      db,
		session: jwtGen,
		hasher:  BcryptHasher{},
		idgen:   failIDGen{},
	}

	t.Run("Generate ID error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)")).
			WithArgs("newuser").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		_, err := repo.Register("newuser", "12345")
		if err == nil {
			t.Errorf("expected error: %v, but got: nil", err)
		}
	})
}
