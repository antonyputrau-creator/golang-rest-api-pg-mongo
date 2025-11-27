package controllers

import (
	"database/sql"
	"log"
	"net/http"
	"regexp"
	"backend-mobile-skill-test/config"
	"backend-mobile-skill-test/models"
	"github.com/gin-gonic/gin"
)

func validateUser(user models.User) (bool, string) {
	if user.Name == "" || user.Email == "" {
		return false, "Semua input (Name dan Email) wajib diisi."
	}

	if !regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`).MatchString(user.Email) {
		return false, "Format email tidak valid."
	}

	if len(user.Name) < 3 || len(user.Name) > 50 {
		return false, "Panjang Nama harus antara 3 dan 50 karakter."
	}

	return true, ""
}

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input body"})
		return
	}

	if ok, msg := validateUser(user); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
	var id int
	err := config.DB.QueryRow(query, user.Name, user.Email).Scan(&id)
	if err != nil {
		log.Println("Database Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user (Email may already exist)"})
		return
	}

	user.ID = id
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}

func GetUserList(c *gin.Context) {
	rows, err := config.DB.Query("SELECT id, name, email FROM users")
	if err != nil {
		log.Println("Database Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			log.Println("Scan Error:", err)
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

func GetUserByID(c *gin.Context) {
	var input struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input body or missing ID"})
		return
	}
	if input.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID wajib diisi"})
		return
	}

	var user models.User
	query := "SELECT id, name, email FROM users WHERE id = $1"
	err := config.DB.QueryRow(query, input.ID).Scan(&user.ID, &user.Name, &user.Email)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		log.Println("Database Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input body"})
		return
	}
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID User wajib di body"})
		return
	}

	if ok, msg := validateUser(user); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	query := "UPDATE users SET name = $1, email = $2 WHERE id = $3"
	result, err := config.DB.Exec(query, user.Name, user.Email, user.ID)
	if err != nil {
		log.Println("Database Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user (Email may already exist)"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found or no changes made"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {
	var input struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input body or missing ID"})
		return
	}
	if input.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID wajib diisi"})
		return
	}

	query := "DELETE FROM users WHERE id = $1"
	result, err := config.DB.Exec(query, input.ID)
	if err != nil {
		log.Println("Database Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}