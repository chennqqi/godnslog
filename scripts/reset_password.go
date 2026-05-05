package main

import (
	"fmt"
	"log"

	"github.com/chennqqi/godnslog/models"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func main() {
	engine, err := xorm.NewEngine("sqlite", "file:godnslog.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer engine.Close()

	// Check if admin user exists
	var user models.TblUser
	has, err := engine.Where("name = ?", "admin").Get(&user)
	if err != nil {
		log.Fatalf("Failed to query user: %v", err)
	}

	if !has {
		fmt.Println("Admin user does not exist, creating...")
		user.Name = "admin"
		user.Email = "admin@example.com"
		user.Pass = makePassword("admin123")
		user.Role = 0 // roleSuper
		user.Lang = "en-US"
		_, err = engine.Insert(&user)
		if err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
		fmt.Println("Admin user created with password: admin123")
	} else {
		fmt.Println("Admin user exists, resetting password...")
		user.Pass = makePassword("admin123")
		_, err = engine.ID(user.Id).Update(&user)
		if err != nil {
			log.Fatalf("Failed to update password: %v", err)
		}
		fmt.Println("Admin password reset to: admin123")
	}
}

func makePassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	return string(hash)
}
