package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	fmt.Println("Desde Data controller")
	c.JSON(200, gin.H{"message": "Get usuario Jesua"})
}
