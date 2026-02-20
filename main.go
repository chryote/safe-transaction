package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func connectDB() {
	var err error

	dsn := "appuser:apppass@tcp(db:3306)/appdb"

	// retry loop (VERY IMPORTANT — containers start at different times)
	for i := 0; i < 20; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil && db.Ping() == nil {
			log.Println("Connected to MySQL!")
			return
		}
		log.Println("Waiting for MySQL...")
		time.Sleep(2 * time.Second)
	}

	log.Fatal("Could not connect to MySQL")
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Payment struct {
	ID         int    `json:"id"`
	ExternalId string `json:"external_id"`
	Amount     int    `json:"amount"`
	Status     string `json:"status"`
}

func main() {
	connectDB()

	r := gin.Default()

	r.GET("/users", getUsers)
	r.POST("/users", createUser)
	r.POST("/payments", createPayment)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	log.Println("Server started")

	// wait for CTRL + C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()

		// Process request
		c.Next()

		// After request finished
		latency := time.Since(start)

		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		log.Printf(
			"[%s] %s %s | %d | %v\n",
			clientIP,
			method,
			path,
			status,
			latency,
		)
	}
}

func getUsers(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		rows.Scan(&u.ID, &u.Name, &u.Email)
		users = append(users, u)
	}
	c.JSON(200, users)
}

func createUser(c *gin.Context) {
	var u User
	c.BindJSON(&u)

	result, err := db.Exec(
		"INSERT INTO users(name,email) VALUES(?,?)",
		u.Name, u.Email,
	)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	u.ID = int(id)

	c.JSON(200, u)
}

func createPayment(c *gin.Context) {
	var p Payment
	if err := c.BindJSON(&p); err != nil {
		c.JSON(400, gin.H{"error": "invalid request."})
		return
	}

	// check if payment already exists
	var existing Payment
	err := db.QueryRow(
		"SELECT id, external_id, amount, status FROM payments where external_id = ?",
		p.ExternalId,
	).Scan(&existing.ID, &existing.ExternalId, &existing.Amount, &existing.Status)

	// always attempt to insert, since external_id on table is UNIQUE. Check the database.
	result, err := db.Exec(
		"INSERT INTO payments(external_id, amount, status) VALUES (?, ?, ?)",
		p.ExternalId,
		p.Amount,
		"SUCCESS",
	)

	if err != nil {

		// if query return an error, it might because of duplicate entry with same external_id
		// search for existing external_id
		var existing Payment

		err2 := db.QueryRow(
			"SELECT id, external_id, amount, status FROM payments WHERE external_id = ?",
			p.ExternalId,
		).Scan(&existing.ID, &existing.ExternalId, &existing.Amount, &existing.Status)

		// found entry(s) return the existing external_id data
		if err2 == nil {
			c.JSON(200, existing)
			return
		}

		c.JSON(500, gin.H{"error": err.Error()})
		return

	}

	id, _ := result.LastInsertId()

	p.ID = int(id)
	p.Status = "SUCCESS"

	c.JSON(200, p)

}
