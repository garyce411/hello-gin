package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseQueryInt(c *gin.Context, key string, defaultVal int) (int, error) {
	query := c.Query(key)
	val, err := strconv.Atoi(query)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		page, err := ParseQueryInt(c, "page", 1)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{
			"message": page,
		})
	})
	r.Run(":8080")
}
