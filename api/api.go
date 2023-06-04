package api

// import (
// 	"database/sql"
// 	"fmt"

// 	_ "github.com/mattn/go-sqlite3"
// )

// var DB *sql.DB

// type User struct {
// 	Id       int
// 	Username string
// 	Mail     string
// 	Password string
// }

// func Init() {
// 	var err error
// 	DB, err = sql.Open("sqlite3", "./bdd.db")
// 	if err != nil {
// 		fmt.Printf("Error opening database: %v\n", err)
// 	}
// }

// func CreateUser(user User) {
// 	querry, _ := DB.Prepare("INSERT INTO users (mail, username, password) VALUES (?, ?, ?)")
// 	querry.Query(user.Mail, user.Username, user.Password)
// 	querry.Close()
// }

// func GetAllUsers() ([]User, error) {
// 	Init()
// 	rows, err := DB.Query("SELECT id, username, mail,password FROM users")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	var users []User
// 	for rows.Next() {
// 		var user User
// 		err = rows.Scan(&user.Mail, &user.Password)
// 		if err != nil {
// 			return nil, err
// 		}
// 		users = append(users, user)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return users, nil
// }
