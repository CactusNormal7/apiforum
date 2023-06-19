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
	Password string `json:"password"`
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
	stmt.Exec(username, mail, password)
	var newUser user
	newUser = user{users[len(users)-1].Id + 1, username, mail, password}
	users = append(users, newUser)

	if username == "" || password == "" || mail == "" {
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		http.Error(c.Writer, "Erreur lors du cryptage du mot de passe", http.StatusInternalServerError)
		return
	}

	stmt2, err := DB.Prepare("INSERT INTO users (username, password, mail) VALUES (?, ?, ?)")
	if err != nil {
		log.Println(err)
		http.Error(c.Writer, "Erreur lors de la préparation de la requête d'insertion", http.StatusInternalServerError)
		return
	}
	defer stmt2.Close()
	tempoTable := []byte{}
	_, err = stmt2.Exec(username, hashedPassword, mail)
	if err != nil {
		log.Println(err)
		http.Error(c.Writer, "Erreur lors de l'insertion de l'utilisateur", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(c.Writer, "L'utilisateur %s a été enregistré avec succès !", username)
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
	stmt, err := DB.Prepare("SELECT * FROM users WHERE username=?")
	if err != nil {
		log.Fatalln(err)
	}
	defer stmt.Close()
	username := c.Query("username")
	var id int
	var name string
	var mail string
	var pswd string
	stmt.QueryRow(username).Scan(&id, &name, &mail, &pswd)
	c.IndentedJSON(http.StatusOK, user{Id: id, Username: name, Mail: mail, Password: pswd})
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
		log.Fatalln(err)
	}
	defer DB.Close()
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS users (ID INTEGER PRIMARY KEY AUTOINCREMENT, USERNAME TEXT, MAIL TEXT, PASSWORD TEXT);")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS channels (ID INTEGER PRIMARY KEY AUTOINCREMENT, ABOUT TEXT);")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = DB.Exec("CREATE TABLE IF NOT EXISTS messages (ID INTEGER PRIMARY KEY AUTOINCREMENT, CONTENT TEXT, SENDERID INTEGER, CHANNELID INTEGER, ISDELETED INTEGER);")
	if err != nil {
		log.Fatalln(err)
	}

	rows, err := DB.Query("SELECT * FROM users")
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var u user
		err = rows.Scan(&u.Id, &u.Username, &u.Mail, &u.Password)
		if err != nil {
			log.Fatalln(err)
		}
		users = append(users, u)
	}

	rows, err = DB.Query("SELECT * FROM channels")
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var c channel
		err = rows.Scan(&c.Id, &c.About)
		if err != nil {
			log.Fatalln(err)
		}
		channels = append(channels, c)
	}

	rows, err = DB.Query("SELECT * FROM messages")
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var m message
		err = rows.Scan(&m.Id, &m.Content, &m.Senderid, &m.Channelid, &m.Isdeleted)
		if err != nil {
			log.Fatalln(err)
		}
		messages = append(messages, m)
	}
}

func main() {
	Init()

	router := gin.Default()

	router.GET("/users", GetUsers)
	router.GET("/messages", GetMessages)
	router.GET("/users/:username", GetUserV)
	router.POST("/users", AddUser)
	router.POST("/messages", AddMsg)
	router.DELETE("/users/:id", RealDeleteUser)
	router.GET("/getmsgsusers/:senderid/:channelid", GetMsgsUsers)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(":"+port, router))
}
