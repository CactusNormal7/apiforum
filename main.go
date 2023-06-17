package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
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
		fmt.Printf("Error opening database: %v\n", err)
	}
}

func ConvertDbUsers(t *testing.T) {
	users = []user{}
	rows, _ := DB.Query("SELECT * FROM users")
	for rows.Next() {
		var ra user
		err := rows.Scan(&ra.Id, &ra.Mail, &ra.Username, &ra.Password)
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
