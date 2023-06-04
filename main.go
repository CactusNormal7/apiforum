package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

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
	stmt.Exec("dauybda", "d&ada", "audaz")

	newUser := user{users[len(users)].Id + 1, "dauybda", "ada", "ada"}
	if err := c.BindJSON(&newUser); err != nil {
		return
	}
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/users", GetUsers)
	router.GET("/messages", GetMessages)
	router.GET("/adduser", AddUser)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Panicf("error: %s", err)
	}
}
