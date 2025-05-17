package utils

import (
	"auth_service/internal/repositiry/storage"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
)

type userData struct {
	Name     string
	Email    string
	PassHash []byte
}

func GenUsers(count int, db *storage.Storage) {
	const batchSize = 10000
	users := make([]userData, 0, batchSize)

	fmt.Println("Starting user generation...")

	for i := 0; i < count/batchSize; i++ {
		fmt.Println("Generated:", i)

		pass := gofakeit.Password(true, true, true, true, false, 8)
		passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("bcrypt error: %v", err)
			continue
		}

		users = append(users, userData{
			Name:     gofakeit.Name(),
			Email:    gofakeit.Email(),
			PassHash: passHash,
		})

		// Когда собрали batch или дошли до конца
		if len(users) == batchSize || i == count-1 {
			if err := insertBatch(db, users); err != nil {
				log.Fatalf("batch insert error: %v", err)
			}
			users = users[:0] // очищаем для следующего батча
		}
	}

	fmt.Println("User generation complete.")
}

func insertBatch(db *storage.Storage, users []userData) error {
	valueStrings := make([]string, 0, len(users))
	valueArgs := make([]interface{}, 0, len(users)*3)

	for i, user := range users {
		idx := i * 3
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", idx+1, idx+2, idx+3))
		valueArgs = append(valueArgs, user.Name, user.Email, user.PassHash)
	}

	stmt := fmt.Sprintf("INSERT INTO users (username, email, passhash) VALUES %s", strings.Join(valueStrings, ","))
	_, err := db.DB.Exec(stmt, valueArgs...)
	return err
}
