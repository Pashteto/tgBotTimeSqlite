package repo

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

type ChatMember struct {
	User  string
	Chat  int64
	Shift int
}

// Execute the SQL statement to create the table if it does not already exist
// where id is the primary key and will be automatically populated
// and name and chatId are text fields and timezone is an integer field
// also creates index on pair of (name, chatId)
func (r *Repo) CreateTableIfNotExists(ctx context.Context) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
    		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    		"name" TEXT,
    		"chat_id" INTEGER,
    		"timezone" INTEGER,
    		UNIQUE(name, chat_id)
      		);`
	_, err := r.db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Table created successfully")
	return nil
}

//	type DbGetterSetter interface {
//		InsertChat(ctx context.Context, id, user string, chat int64, shift int) error
//		DelUser(ctx context.Context, chat int64, user string)
//		GetCount(ctx context.Context, chat int64) (int, error)
//		GetChat(ctx context.Context, chat int64) ([]repo.ChatMember, error)
//	}
type DbGetterSetter interface {
	InsertTimezoneUserChat(tz int, name string, chat int64) error
	DelUser(chat int64, user string)
	ListChatUsers(chat int64) ([]ChatMember, error)
	FindTimeZoneByChatIdAndName(name string, chat int64) (int, error)
}

// InsertTimezoneUserChat inserts a new record into the sqlite table users.
// if record exists - it will be replaced
func (r *Repo) InsertTimezoneUserChat(tz int, name string, chat int64) error {
	// Prepare the SQL statement for execution
	stmt, err := r.db.Prepare("INSERT OR REPLACE INTO users(name, chat_id, timezone) VALUES(?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the SQL statement
	_, err = stmt.Exec(name, chat, tz)
	if err != nil {
		return err
	}

	return nil
}

// FindTimeZoneByChatIdAndName finds a timezone by chatId and name
func (r *Repo) FindTimeZoneByChatIdAndName(name string, chat int64) (int, error) {
	// Prepare the SQL statement for execution
	stmt, err := r.db.Prepare("SELECT timezone FROM users WHERE name = ? AND chat_id = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	// Query the row returned by the SQL statement
	var tz int
	err = stmt.QueryRow(name, chat).Scan(&tz)
	if err != nil {
		return 0, err
	}
	return tz, nil
}

// DelUser deletes a record from the database.
func (r *Repo) DelUser(chat int64, user string) {
	r.db.QueryRow("DELETE FROM users WHERE chat_id = ? AND name = ?", chat, user)
}

// ListChatUsers finds all users by chatId
func (r *Repo) ListChatUsers(chat int64) ([]ChatMember, error) {
	rows, err := r.db.Query("SELECT name, chat_id, timezone FROM users WHERE chat_id = ?", chat)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var users []ChatMember
	for rows.Next() {
		var user ChatMember
		err = rows.Scan(&user.User, &user.Chat, &user.Shift)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetCount returns a list of users in the chat.
func (r *Repo) GetCount(chat int64) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE chat_id = ?", chat).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

/*
// InsertChat inserts a into the database.
func (r *Repo) InsertChat(ctx context.Context, id, user string, chat int64, shift int) error {
	var count int

	err := sqldb.QueryRow(context.TODO(), `
		SELECT COUNT(*)
		FROM url
		WHERE chat = $1
		AND username = $2
		`, chat, user).Scan(&count)
	if err != nil {
		return fmt.Errorf("insertChat: count: %w", err)
	}

	if count == 0 {
		fakeurl := "ya.ru"
		_, err = sqldb.Exec(context.TODO(), `
			INSERT INTO url (id, original_url, chat, username, shift)
			VALUES ($1, $2, $3, $4, $5)
		`, id, fakeurl, chat, user, shift)
		if err != nil {
			return fmt.Errorf("insertChat: count == 0: %w", err)
		}
		return nil
	} else {
		fmt.Println(id)

		err = sqldb.QueryRow(context.TODO(), `
			SELECT id FROM url
			WHERE chat = $1
			AND username = $2
		`, chat, user).Scan(&id)
		if err != nil {
			return fmt.Errorf("insertChat: count != 0: %w", err)
		}
		fmt.Println(id)

		_, err = sqldb.Exec(context.TODO(), `
			UPDATE url
			SET shift = $1
			WHERE id = $2
		`, shift, id)
		if err != nil {
			return fmt.Errorf("insertChat: update: %w", err)
		}
		return nil
	}
}

// DelUser deletes a record from the database.
func (r *Repo) DelUser(ctx context.Context, chat int64, user string) {
	sqldb.QueryRow(context.TODO(), `
		DELETE FROM url
		WHERE chat = $1
  		AND username = $2
  `, chat, user)
}

// GetCount counts users in db.
func (r *Repo) GetCount(ctx context.Context, chat int64) (int, error) {
	var count int
	err := sqldb.QueryRow(context.TODO(), `
		SELECT COUNT(*)
		FROM url
		WHERE chat = $1
		`, chat).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetChat lists users in db.
func (r *Repo) GetChat(ctx context.Context, chat int64) ([]ChatMember, error) {
	rows, err := sqldb.Query(context.TODO(), `
		SELECT username, shift FROM url
		WHERE chat = $1
	`, chat)
	defer rows.Close()

	var people []ChatMember

	for rows.Next() {
		var user string
		var shift int

		if err := rows.Scan(&user, &shift); err != nil {
			return nil, fmt.Errorf("dsfkvblkdfjfbvb %w", err)
		}

		p := ChatMember{
			User:  user,
			Chat:  chat,
			Shift: shift,
		}
		people = append(people, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return people, err
}
*/
