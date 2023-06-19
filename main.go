package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"

)

var DB *sql.DB
var store = sessions.NewCookieStore([]byte("super-secret"))

type user struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Mail     string `json:"mail"`
	HashedPassword string `json:"hashed_password"`
}

type message struct {
	Id        int    `json:"id"`
	Content   string `json:"content"`
	Senderid  int    `json:"senderid"`
	Channelid int    `json:"channelid"`
	Isdeleted int    `json:"isdeleted"`
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
	stmt, err := DB.Prepare("INSERT INTO users (username, mail, password) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	username := c.Query("username")
	password := c.Query("password")
	mail := c.Query("mail")

	// Hashage du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalln(err)
	}

	stmt.Exec(username, mail, hashedPassword)
	var newUser user
	newUser = user{users[len(users)-1].Id + 1, username, mail, string(hashedPassword)}
	users = append(users, newUser)

	if username == "" || mail == "" || password == "" {
		http.Error(c.Writer, "Veuillez remplir tous les champs du formulaire d'inscription", http.StatusBadRequest)
		return
	}

	if !isStrongPassword(password) {
		http.Error(c.Writer, "Le mot de passe doit contenir au moins 8 caractères, dont au moins une lettre majuscule, un chiffre et un caractère spécial", http.StatusBadRequest)
		return
	}

	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		log.Println(err)
		http.Error(c.Writer, "Erreur lors de la vérification du nom d'utilisateur", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(c.Writer, "Le nom d'utilisateur est déjà utilisé", http.StatusBadRequest)
		return
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE mail = ?", mail).Scan(&count)
	if err != nil {
		log.Println(err)
		http.Error(c.Writer, "Erreur lors de la vérification de l'adresse e-mail", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(c.Writer, "L'adresse e-mail est déjà utilisée", http.StatusBadRequest)
		return
	}

}


func RealDeleteUser(c *gin.Context) {
	stmt, err := DB.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	idtodel := c.Query("id")
	idToDelete, _ := strconv.Atoi(idtodel)
	stmt.Exec(idToDelete)
	ConvertDbUsers(&testing.T{})
}

func GetUserV(c *gin.Context) {
	stmt, err := DB.Prepare("SELECT password FROM users WHERE username=?")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	username := c.Query("username")
	var storedPassword string
	err = stmt.QueryRow(username).Scan(&storedPassword)
	if err != nil {
		log.Fatalln(err)
	}

	password := c.Query("password")

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		fmt.Fprintln(c.Writer, "Nom d'utilisateur ou mot de passe invalide !")
		return
	}

	// Si la comparaison réussit, l'authentification est valide
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Authentification réussie"})
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

func AddMsg(c *gin.Context) {
	stmt, err := DB.Prepare("INSERT INTO messages (CONTENT, SENDERID, CHANNELID, ISDELETED ) VALUES (?, ?, ?, 0)")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	content := c.Query("content")
	senderid, _ := strconv.Atoi(c.Query("senderid"))
	channelid, _ := strconv.Atoi(c.Query("channelid"))
	stmt.Exec(content, senderid, channelid)
	newMsg := message{messages[len(messages)-1].Id + 1, content, senderid, channelid, 0}
	messages = append(messages, newMsg)
}

func GetMsgsUsers(c *gin.Context) {
	senderid := c.Query("senderid")
	channelid := c.Query("channelid")
	newmsg := []message{}
	rows, _ := DB.Query("SELECT * FROM messages WHERE SENDERID=? and CHANNELID=?", senderid, channelid)
	for rows.Next() {
		var ra message
		err := rows.Scan(&ra.Id, &ra.Content, &ra.Senderid, &ra.Channelid, &ra.Isdeleted)
		newmsg = append(newmsg, ra)
		if err != nil {
			log.Fatalln(err)
		}
	}
	c.IndentedJSON(http.StatusOK, newmsg)
}

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "./bdd.db")
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
	}
}

func ConvertDbUsers(t *testing.T) {
	users = []user{}
	rows, _ := DB.Query("SELECT * FROM users")
	for rows.Next() {
		var ra user
		err := rows.Scan(&ra.Id, &ra.Mail, &ra.Username, &ra.HashedPassword)
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
		err := rows.Scan(&ra.Id, &ra.Content, &ra.Senderid, &ra.Channelid, &ra.Isdeleted)
		messages = append(messages, ra)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func main() {
	Init()
	ConvertDbUsers(&testing.T{})
	ConvertMsg(&testing.T{})
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/users", GetUsers)
	router.GET("/messages", GetMessages)
	router.GET("/adduser", AddUser)
	router.GET("/rdeleteuser", RealDeleteUser)
	router.GET("/getuserv", GetUserV)
	router.GET("/addmsg", AddMsg)
	router.GET("/getmsgs", GetMsgsUsers)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Panicf("error: %s", err)
	}
}
