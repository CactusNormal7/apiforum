package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"testing"
	"unicode"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

type user struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type message struct {
	Id        int    `json:"id"`
	Content   string `json:"content"`
	SenderID  int    `json:"senderid"`
	ChannelID int    `json:"channelid"`
	IsDeleted int    `json:"isdeleted"`
}

type channel struct {
	Id    int    `json:"id"`
	About string `json:"about"`
}

var users = []user{}
var messages = []message{}
var channels = []channel{}

func GetUsers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, users)
}

func GetMessages(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, messages)
}

func AddUser(c *gin.Context) {
	stmt, err := DB.Prepare("INSERT INTO users (username, email, password) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()

	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if username == "" || email == "" || password == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if !isStrongPassword(password) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if count > 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if count > 0 {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(username, email, hashedPassword)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	newUser := user{
		Id:       len(users) + 1,
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}
	users = append(users, newUser)

	fmt.Fprintf(c.Writer, "L'utilisateur %s a été enregistré avec succès !", username)
}

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "./bdd.db")
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
	}
}

func ConvertDbUsers(t *testing.T) {
	rows, _ := DB.Query("SELECT * FROM users")
	for rows.Next() {
		var ra user
		err := rows.Scan(&ra.Id, &ra.Username, &ra.Email, &ra.Password)
		users = append(users, ra)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func ConvertMsg(t *testing.T) {
	rows, _ := DB.Query("SELECT * FROM messages")
	for rows.Next() {
		var ra message
		err := rows.Scan(&ra.Id, &ra.Content, &ra.SenderID, &ra.ChannelID, &ra.IsDeleted)
		messages = append(messages, ra)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	switch c.Request.Method {
	case http.MethodGet:
		fmt.Fprintln(c.Writer, "Veuillez vous connecter !")
	case http.MethodPost:
		var (
			storedUsername string
			storedPassword string
		)
		err := DB.QueryRow("SELECT username, password FROM users WHERE username = ?", username).Scan(&storedUsername, &storedPassword)
		if err == sql.ErrNoRows {
			fmt.Fprintln(c.Writer, "Nom d'utilisateur ou mot de passe invalide !")
			return
		}
		logError(err)
		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		if err != nil {
			fmt.Fprintln(c.Writer, "Nom d'utilisateur ou mot de passe invalide !")
			return
		}
	}
}

func isStrongPassword(password string) bool {
	const (
		minLength    = 8
		minDigits    = 1
		minSymbols   = 1
		minUppercase = 1
	)

	if len(password) < minLength {
		return false
	}

	var digits, symbols, uppercase int
	for _, char := range password {
		switch {
		case unicode.IsDigit(char):
			digits++
		case unicode.IsSymbol(char) || unicode.IsPunct(char):
			symbols++
		case unicode.IsUpper(char):
			uppercase++
		}
	}

	if digits < minDigits || symbols < minSymbols || uppercase < minUppercase {
		return false
	}

	return true
}

func logError(err error) {
	if err != nil {
		panic(err)
	}
}


// func main() {
// 	Init()

// 	router := gin.Default()

// 	router.GET("/users", GetUsers)
// 	router.GET("/messages", GetMessages)
// 	router.POST("/users", AddUser)
// 	router.POST("/login", Login)


// 	err := router.Run(":8080")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }