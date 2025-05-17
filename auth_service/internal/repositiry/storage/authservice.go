package storage

import (
	"auth_service/internal/config"
	"auth_service/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log/slog"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExist    = errors.New("user already exists")
)

type Storage struct {
	DB *sqlx.DB
}

func New(config *config.Config) (*Storage, error) {
	cfg := config.StorageCfg

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", cfg.Host, cfg.Username, cfg.Password, cfg.DBName, cfg.Port)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	if _, err := db.Conn(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	slog.Info("storage run")
	fmt.Println(dsn)
	return &Storage{DB: db}, nil
}

func (s *Storage) GetUserById(userId int64) (models.User, error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return models.User{}, fmt.Errorf("error begin transaction %v", err)
	}

	defer func() {
		if err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				return
			}
			return
		}

		err = tx.Commit()
	}()

	return s.getUserById(tx, userId)
}

func (s *Storage) getUserById(tx *sqlx.Tx, userId int64) (models.User, error) {
	var user models.User
	err := tx.Get(&user, "SELECT * FROM users WHERE id=$1", userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}

	return user, nil
}

func (s *Storage) GetUser(username string) (user models.User, err error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return models.User{}, fmt.Errorf("error begin transaction %v", err)
	}

	defer func() {
		if err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				return
			}
			return
		}

		err = tx.Commit()
	}()

	return s.getUser(tx, username)
}

func (s *Storage) getUser(tx *sqlx.Tx, username string) (models.User, error) {
	var user models.User
	err := tx.Get(&user, "SELECT * FROM users WHERE username=$1", username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}

	return user, nil
}

func (s *Storage) AddNewUser(newUser models.User) (id int64, err error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return -1, fmt.Errorf("error begin transaction %v", err)
	}

	defer func() {
		if err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				return
			}
			return
		}
		err = tx.Commit()
	}()

	return s.addNewUser(tx, newUser)

}

func (s *Storage) addNewUser(tx *sqlx.Tx, newUser models.User) (id int64, err error) {
	var exists bool
	err = tx.Get(&exists, `
    SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)`,
		newUser.Username, newUser.Email,
	)
	if err != nil {
		return -1, fmt.Errorf("select exists error: %w", err)
	}
	if exists {
		return -1, fmt.Errorf("user already exists")
	}

	_, err = tx.Exec("INSERT INTO users (email, username, passhash) VALUES ($1, $2, $3)", newUser.Email, newUser.Username, newUser.PassHash)
	if err != nil {
		return -1, err
	}

	u, err := s.getUser(tx, newUser.Username)
	if err != nil {
		slog.Warn("cant get user from added user method")
		return -1, err
	}

	return u.ID, nil
}

func (s *Storage) RemoveUser(username string) error {
	_, err := s.DB.Exec("DELETE FROM users WHERE username=$1", username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *Storage) UserRole(username string) (string, error) {
	var role string
	err := s.DB.Get(&role, "SELECT role FROM users WHERE username=$1", username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return role, nil
}

func (s *Storage) IsAdmin(userId int64) (bool, error) {
	type req struct {
		Username string `db:"username"`
	}
	var user req
	err := s.DB.Get(&user, "SELECT username FROM admins WHERE id = $1", userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("user is not an admin")
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *Storage) CheckUser(username, email string) error {
	var user models.User
	err := s.DB.Get(&user, "SELECT * FROM users WHERE username = $1 or email = $2", username, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return ErrUserExist
	}

	if user.ID > 0 {
		return ErrUserExist
	}

	return nil
}
