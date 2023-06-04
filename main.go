package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
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

var users = []user{}

func GetUsers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, users)
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

func main() {
	Init()
	ConvertDbUsers(&testing.T{})
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/users", GetUsers)
	router.Run("localhost:8080")
}
