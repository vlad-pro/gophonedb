package main

import (
	"fmt"
	"regexp"

	_ "github.com/lib/pq"
	phonedb "github.com/vlad-pro/gophonedb/db"
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
	must(phonedb.Reset("postgres", psqlInfo, dbname))

	// TODO: find out why using the first expression is not working properly. No database with dbname at first.
	psqlInfo = "postgres://gopher:gopher@localhost/gophercises_phone?sslmode=disable"
	must(phonedb.Migrate("postgres", psqlInfo))

	db, err := phonedb.Open("postgres", psqlInfo)
	must(err)
	defer db.Close()

	must(db.Seed())

	// must(db.Ping()) just a usefull function to know about. It has a similar function in mongoDB

	phones, err := db.AllPhones()
	must(err)
	for _, p := range phones {
		fmt.Printf("Working on%+v\n", p)
		number := normalize(p.Number)
		if number != p.Number {
			fmt.Println("Updating or removing...", number)
			existing, err := db.FindPhone(number)
			must(err)
			if existing != nil {
				must(db.DeletePhone(p.ID))
			} else {
				p.Number = number
				must(db.UpdatePhone(&p))
			}
		} else {
			fmt.Println("No changes required")
		}
	}

	// id, err := insertPhone(db, "861234567890")
	// must(err)
	// fmt.Println("id=", id)
	println("done")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
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
