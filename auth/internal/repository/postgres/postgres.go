package postgres

import (
	"auth/internal/config"
	"auth/internal/models"
	"auth/pkg/hash"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

type Storage struct {
	DB *sqlx.DB
}

func New(cfg *config.Config) (*Storage, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBCfg.Host,
		cfg.DBCfg.Port,
		cfg.DBCfg.Username,
		cfg.DBCfg.Password,
		cfg.DBCfg.DBName,
	)
	// подсоединяет к бд и делает пинг, чтобы убедится что все ок
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	log.Println("connected to db")

	return &Storage{DB: db}, nil
}

func (db *Storage) GetUserByEmail(email string) models.User {
	query := `SELECT * FROM users WHERE email=$1`

	var user models.User
	if err := db.DB.Get(&user, query, email); err != nil {
		log.Println(err)
	}

	return user
}

// CreateUser создает нового пользователя и возвращает его id
func (db *Storage) CreateUser(user models.User) (int, error) {
	query := `INSERT INTO users (email, passhash, username) VALUES ($1, $2, $3) RETURNING id`

	var id int
	err := db.DB.QueryRow(query, user.Email, user.PassHash, user.Username).Scan(&id)
	if err != nil {
		log.Println(err)
		return -1, err
	}

	return id, nil
}

func (db *Storage) UserExists(email, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1 OR username=$2)`

	var exists bool
	err := db.DB.Get(&exists, query, email, username)
	if err != nil {
		log.Println("Error checking user existence:", err)
		return false, err
	}

	return exists, nil
}

func (db *Storage) VerifyUserWithCredentials(username, password string) (models.User, error) {
	query := `SELECT * FROM users WHERE username=$1 LIMIT 1`

	var user models.User
	err := db.DB.Get(&user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("User not found")
			return models.User{}, errors.New("invalid credentials")
		}
		log.Println(err)
		return models.User{}, err
	}

	isEqual := hash.ComparePasswords(user.PassHash, password)
	if !isEqual {
		log.Println("Invalid password")
		return models.User{}, errors.New("invalid credentials")
	}

	return user, nil
}
