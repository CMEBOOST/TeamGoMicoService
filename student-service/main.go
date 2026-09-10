package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// สร้าง Struct สำหรับข้อมูลนักศึกษา
type Student struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `json:"name" binding:"required"`
	StudentID string `json:"student_id" binding:"required"`
}

var db *gorm.DB

func initDB() {
	dsn := "host=postgres-db user=postgres password=secret dbname=userdb port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	db.AutoMigrate(&Student{})
}

func main() {
	time.Sleep(5 * time.Second)
	initDB()

	r := gin.Default()

	// สร้างนักศึกษาใหม่
	r.POST("/students", func(c *gin.Context) {
		var student Student
		if err := c.ShouldBindJSON(&student); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Create(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create student"})
			return
		}
		c.JSON(http.StatusCreated, student)
	})

	// ดึงข้อมูลนักศึกษาทั้งหมด
	r.GET("/students", func(c *gin.Context) {
		var students []Student
		db.Find(&students)
		c.JSON(http.StatusOK, students)
	})

	// รัน Service นี้ที่พอร์ต 8081
	r.Run(":8081")
}
