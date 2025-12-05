package main

import (
	"database/sql"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/postgres"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type User struct {
	Id     uint
	Uname  string
	Passwd string
}

type Users struct {
	Users []User
}

func main() {

	router := gin.Default()
	db, err := sql.Open("postgres", "postgresql://azz:roxy@localhost:5432/roxer")
	if err != nil {
		panic(err)
	}
	store, err := postgres.NewStore(db, []byte("secret"))
	if err != nil {
		panic(err)
	}

	router.Use(sessions.Sessions("dash", store))
	router.LoadHTMLGlob("views/html/*")
	router.Static("/style", "/home/azz/Projects/go/Roxer/views/style")
	router.Use(sessions.Sessions("mySession", store))
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Roxer",
		})
	})

	router.GET("register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{})
	})

	router.GET("login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})

	router.POST("/register", func(c *gin.Context) {
		var user User
		if err := c.ShouldBind(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
			return
		}

		// Hash the password
		hashed, err := bcrypt.GenerateFromPassword([]byte(user.Passwd), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Insert into database
		_, err = db.Exec("INSERT INTO users (uname, password) VALUES ($1, $2)", user.Uname, string(hashed))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/login")
	})

	// POST /login - authenticate existing user
	router.POST("/login", func(c *gin.Context) {
		var loginUser User
		if err := c.ShouldBind(&loginUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
			return
		}

		// Get user from DB
		var stored User
		err := db.QueryRow("SELECT id, uname, password FROM users WHERE uname = $1", loginUser.Uname).
			Scan(&stored.Id, &stored.Uname, &stored.Passwd)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Compare passwords
		if bcrypt.CompareHashAndPassword([]byte(stored.Passwd), []byte(loginUser.Passwd)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Set session
		session := sessions.Default(c)
		session.Set("user", stored.Uname)
		session.Save()

		c.Redirect(http.StatusFound, "/dashboard")
	})

	// Dashboard (requires login)
	router.GET("/api/isLoggedIn", func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"loggedIn": false})
		} else {
			c.JSON(http.StatusOK, gin.H{"loggedIn": true, "user": user})
		}
	})

	router.GET("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(http.StatusFound, "/login")
	})

	router.Run(":3000")
}
