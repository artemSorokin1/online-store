package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Product struct {
	ID          int           `db:"id"`
	Name        string        `db:"name"`
	Price       int           `db:"price"`
	Sizes       pq.Int64Array `db:"sizes"`
	ImageURL    string        `db:"imageurl"`
	Description string        `db:"description"`
}

var words = []string{"the", "be", "to", "of", "and", "a", "in", "that", "have", "I",
	"it", "for", "not", "on", "with", "he", "as", "you", "do", "at",
	"this", "but", "his", "by", "from", "they", "we", "say", "her", "she",
	"or", "an", "will", "my", "one", "all", "would", "there", "their", "what",
	"so", "up", "out", "if", "about", "who", "get", "which", "go", "me",
	"when", "make", "can", "like", "time", "no", "just", "him", "know", "take",
	"people", "into", "year", "your", "good", "some", "could", "them", "see", "other",
	"than", "then", "now", "look", "only", "come", "its", "over", "think", "also",
	"back", "after", "use", "two", "how", "our", "work", "first", "well", "way",
	"even", "new", "want", "because", "any", "these", "give", "day", "most", "us"}

func GenProducts(cnt int, db *sqlx.DB) {
	var wg sync.WaitGroup

	chunkSize := cnt / runtime.NumCPU()
	for j := 0; j < runtime.NumCPU(); j++ {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			tx := db.MustBegin()
			for i := start; i < end; i++ {
				cntWords := rand.Intn(5) + 15
				randSent := make([]string, 0, cntWords)

				if i%100000 == 0 {
					fmt.Println("Generating product: ", i)
				}

				var sb strings.Builder

				strI := strconv.Itoa(i)

				sb.WriteString("Product_")
				sb.WriteString(strI)

				var sizes = pq.Int64Array([]int64{37, 38, 39})

				for range cntWords {
					randSent = append(randSent, words[rand.Intn(len(words))])
				}

				var product = Product{
					Name:        sb.String(),
					Price:       100,
					Sizes:       sizes,
					ImageURL:    "https://some_url",
					Description: strings.Join(randSent, " "),
				}
				q := "INSERT INTO products (name, price, sizes, imageurl, description) VALUES ($1, $2, $3, $4, $5)"
				tx.Exec(q, product.Name, product.Price, product.Sizes, product.ImageURL, product.Description)
			}
			tx.Commit()
		}(j*chunkSize, (j+1)*chunkSize)
	}

	wg.Wait()
}

const (
	CntProducts = 5_000_000
)

func NewDbConn() (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		"localhost",
		"root",
		"123",
		"delivery",
		"5433")

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := db.Conn(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}
	return db, nil
}

func main() {
	db, err := NewDbConn()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	GenProducts(CntProducts, db)
}
