package models

import (
	"github.com/lightswitch/dutchman-backend/dutchman/util/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id" bson:"id"`
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}

func NewUser(email string, name string, password string) *User {
	uid, _ := uuid.NewV4()

	pwd, _ := hashPassword(password)
	return &User{
		ID:       uid.String(),
		Name:     name,
		Email:    email,
		Password: pwd,
	}
}

func (u *User) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}
