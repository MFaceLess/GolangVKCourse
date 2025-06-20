package session

import (
	"database/sql"
	"time"
)

type SessionMySQLRepository struct {
	db *sql.DB
}

func NewSessionMySQLRepo(db *sql.DB) *SessionMySQLRepository {
	return &SessionMySQLRepository{db: db}
}

func (repo *SessionMySQLRepository) DeleteSession(login string) (int64, error) {
	result, err := repo.db.Exec(
		"DELETE FROM sessions WHERE login = ?",
		login,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (repo *SessionMySQLRepository) UpsertSession(jwtStr, login string, expiredAt time.Time) (int64, error) {
	result, err := repo.db.Exec(
		"INSERT INTO sessions (`login`, `jwt`, `expires_at`) VALUES (?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE `jwt` = VALUES(`jwt`), `expires_at` = VALUES(`expires_at`)",
		login,
		jwtStr,
		expiredAt,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (repo *SessionMySQLRepository) ValidateToken(login, jwt string) (bool, error) {
	var storedJWT string
	err := repo.db.QueryRow(
		"SELECT jwt FROM sessions WHERE login = ?",
		login,
	).Scan(&storedJWT)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}
	return storedJWT == jwt, nil
}
