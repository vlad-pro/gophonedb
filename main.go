package main

import (
	"database/sql"
	"fmt"
	"regexp"

	phonedb "github.com/gophercises/normalize/db"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "gopher"
	password = "gopher"
	dbname   = "gophercises_phone"
	// user = "postgres"

	// password = ""
)

func main() {
	// this line from TODO:
	// psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)
	psqlInfo := "postgres://gopher:gopher@localhost/postgres?sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	must(err)
	err = resetDB(db, dbname)
	must(err)
	db.Close()
	must(phonedb.Reset("postgres", psqlInfo, dbname))

	// TODO: find out why using the first expression is not working properly. No database with dbname at first.
	psqlInfo = "postgres://gopher:gopher@localhost/gophercises_phone?sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	must(err)

	defer db.Close()

	// must(db.Ping()) just a usefull function to know about. It has a similar function in mongoDB

	must(createPhoneNumbersTable(db))

	_, err = insertPhone(db, "1234567890")
	must(err)
	_, err = insertPhone(db, "123 456 7891")
	must(err)
	id, err := insertPhone(db, "(123) 456 7892")
	must(err)
	_, err = insertPhone(db, "(123) 456-7893")
	must(err)
	_, err = insertPhone(db, "123-456-7894")
	must(err)
	_, err = insertPhone(db, "123-456-7890")
	must(err)
	_, err = insertPhone(db, "1234567892")
	must(err)
	_, err = insertPhone(db, "(123)456-7892")
	must(err)

	number, err := getPhone(db, id)
	must(err)
	fmt.Println("Number is ... ", number)

	phones, err := allPhones(db)
	must(err)
	for _, p := range phones {
		fmt.Printf("Working on%+v\n", p)
		number := normalize(p.number)
		if number != p.number {
			fmt.Println("Updating or removing...", number)
			existing, err := findPhone(db, number)
			must(err)
			if existing != nil {
				must(deletePhone(db, p.id))
			} else {
				p.number = number
				must(updatePhone(db, p))
			}
		} else {
			fmt.Println("No changes required")
		}
	}

	// id, err := insertPhone(db, "861234567890")
	// must(err)
	// fmt.Println("id=", id)
}

func getPhone(db *sql.DB, id int) (string, error) {
	var number string
	err := db.QueryRow("SELECT * FROM phone_numbers WHERE id=$1", id).Scan(&id, &number)
	if err != nil {
		return "", err
	}
	return number, nil
}

func findPhone(db *sql.DB, number string) (*phone, error) {
	var p phone
	err := db.QueryRow("SELECT * FROM phone_numbers WHERE value=$1", number).Scan(&p.id, &p.number)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return &p, nil
}

func updatePhone(db *sql.DB, p phone) error {
	statement := `UPDATE phone_numbers SET value=$2 WHERE id=$1`
	_, err := db.Exec(statement, p.id, p.number)
	return err
}

func deletePhone(db *sql.DB, id int) error {
	statement := `DELETE FROM phone_numbers WHERE id=$1`
	_, err := db.Exec(statement, id)
	return err
}

type phone struct {
	id     int
	number string
}

func allPhones(db *sql.DB) ([]phone, error) {
	rows, err := db.Query("SELECT id, value FROM phone_numbers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ret []phone

	for rows.Next() {
		var p phone
		if err := rows.Scan(&p.id, &p.number); err != nil {
			return nil, err
		}
		ret = append(ret, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func insertPhone(db *sql.DB, phone string) (int, error) {
	statement := `INSERT INTO PHONE_NUMBERS(value) VALUES($1) RETURNING id`
	var id int
	err := db.QueryRow(statement, phone).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func createPhoneNumbersTable(db *sql.DB) error {
	statement := `
	CREATE TABLE IF NOT EXISTS phone_numbers (
		id SERIAL,
		value VARCHAR(255)
	)
	`
	_, err := db.Exec(statement)
	return err
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func resetDB(db *sql.DB, name string) error {
	_, err := db.Exec("DROP DATABASE IF EXISTS " + name)
	if err != nil {
		panic(err)
	}
	return createDB(db, name)
}

func createDB(db *sql.DB, name string) error {
	_, err := db.Exec("CREATE DATABASE " + name)
	if err != nil {
		panic(err)
	}
	return nil
}

func normalize(phone string) string {
	//  re := regexp.MustCompile("[^0-9]")
	re := regexp.MustCompile("\\D")
	return re.ReplaceAllString(phone, "")
}

// func normalize(phone string) string {
// 	var buf bytes.Buffer // faster than adding to a empty string result := "" + s
// 	for _, ch := range phone {
// 		if ch >= '0' && ch <= '9' {
// 			buf.WriteRune(ch)
// 		}
// 	}
// 	return buf.String()
// }
