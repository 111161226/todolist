package db

// schema.go provides data models in DB
import (
	"time"
)

// Task corresponds to a row in `tasks` table
type Task struct {
	ID        uint64    `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	IsDone    bool      `db:"is_done"`
	Priority  uint64    `db:"priority"`
}

// User corresponds to 
type User struct {
    ID        uint64    `db:"id"`
    Name      string    `db:"name"`
    Password  []byte    `db:"password"`
	Updated_at time.Time `db:"updated_at"`
}