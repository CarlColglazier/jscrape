package main
import (
	"fmt"
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	last_id = 0
)

type Database struct {
	db *sql.Tx
	insert_category *sql.Stmt
	insert_clue *sql.Stmt
	insert_classification *sql.Stmt
}

func CreateDatabase(path string) (*Database, error) {
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	prepare_database(database)
	db, err := database.Begin()
	if err != nil {
		return nil, err
	}
	category_statement, err := db.Prepare(
		"INSERT OR IGNORE INTO categories(category, description) VALUES(?, ?);")
	if err != nil {
		return nil, err
	}
	insert_clue, err := db.Prepare(
		"INSERT INTO clues(clue, answer, position, round) VALUES(?, ?, ?, ?);")
	if err != nil {
		return nil, err
	}
	classifications_statement, err := db.Prepare(
		"INSERT INTO classifications VALUES(?, ?, ?);")
	if err != nil {
		return nil, err
	}
	prepared_database := Database{
		db: db,
		insert_category: category_statement,
		insert_clue: insert_clue,
		insert_classification: classifications_statement,
	}
	return &prepared_database, nil
}

func (database *Database) Exec(query string, args ...interface{}) {
	_, err := database.db.Exec(query)
	if err != nil {
		fmt.Println(err)
	}
}

func prepare_database(database *sql.DB) {
	fmt.Println("Preparing database")
	database.Exec("PRAGMA foreign_keys = ON;")

	// Create games
	database.Exec(`CREATE TABLE IF NOT EXISTS games(
            id INTEGER PRIMARY KEY,
            airdate TEXT
        );`)
	// Create clues
	database.Exec(`CREATE TABLE IF NOT EXISTS clues(
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            clue TEXT,
            answer TEXT,
            position INTEGER,
            round INTEGER
        );`)
	database.Exec(`CREATE TABLE IF NOT EXISTS categories(
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            category TEXT UNIQUE,
            description TEXT
        );`)
	database.Exec(`CREATE TABLE IF NOT EXISTS classifications(
            clue_id INTEGER,
            category_id INTEGER,
            game_id INTEGER,
            FOREIGN KEY(clue_id) REFERENCES clues(id),
            FOREIGN KEY(category_id) REFERENCES categories(id),
            FOREIGN KEY(game_id) REFERENCES games(id)
        );`)
}

func (database *Database) InsertGame(game Game) {
	_, err := database.db.Exec(
		`INSERT OR IGNORE INTO games(id, airdate) VALUES(?, ?);`,
		game.id,
		game.airdate)
	if err != nil {
		log.Fatal(err)
	}
	for _, category := range game.categories {
		_, err = database.InsertCategory(category, game.id)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (database *Database) InsertCategory(category Category, game_id int) (id int, err error) {
	if _, err = database.insert_category.Exec(category.name, category.description); err != nil {
		fmt.Println("Could not create category.")
		return
	}
	if err = database.db.QueryRow(
		`SELECT id FROM categories WHERE category=?;`,
		category.name).Scan(&id); err != nil {
			return
		}

	for _, clue := range category.clues {
		clue_id, err := database.InsertClue(clue)
		if err != nil {
			log.Fatal(err)
		}
		database.insert_classification.Exec(
			clue_id,
			id,
			game_id)
	}
	return
}

func (database *Database) InsertClue(clue Clue) (id int, err error) {
	_, err = database.insert_clue.Exec(
		clue.clue,
		clue.answer,
		clue.position,
		clue.round)
	if err != nil {
		fmt.Println(err)
	}
	last_id++
	id = last_id
	return
}
