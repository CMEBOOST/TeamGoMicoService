package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique;not null" json:"username" binding:"required"`
	Password string `gorm:"not null" json:"password" binding:"required"`
}

var db *gorm.DB
var jwtSecret = []byte("my-super-secret-key") // คีย์ลับสำหรับสร้าง JWT

func initDB() {
	dsn := "host=postgres-db user=postgres password=secret dbname=userdb port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database")
	}
	db.AutoMigrate(&User{})
}

func main() {
	// หน่วงเวลาเล็กน้อย รอให้ Postgres ใน Docker เปิดเสร็จก่อน
	time.Sleep(5 * time.Second)
	initDB()

	r := gin.Default()

	// Endpoint สมัครสมาชิก
	r.POST("/register", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// นำรหัสผ่านไป Hash ก่อนบันทึกลงฐานข้อมูล
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		user.Password = string(hashedPassword)

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Username already exists"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	})

	// Endpoint ล็อกอิน
	r.POST("/login", func(c *gin.Context) {
		var input User
		var user User

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// หา User จาก Username
		if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// ตรวจสอบรหัสผ่านที่ Hash ไว้
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// สร้าง JWT Token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss":     "my-issuer-key",
			"user_id": user.ID,
			"exp":     time.Now().Add(time.Hour * 2).Unix(), // หมดอายุใน 2 ชม.
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	})

	r.Run(":8080")
}
