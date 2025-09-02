package main

import (
	//"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	router := gin.Default()
	router.LoadHTMLFiles("views/index.html")
	router.Static("/style", "/home/azz/Projects/go/Roxer/views/style/")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Roxer",
		})
	})
	router.Run(":3000")
}
