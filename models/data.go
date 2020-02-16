package models

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"time"
)

type DataStorer interface {
	SaveUser(user User) (AuthResponse, error)
	FetchUser(username string) (User, error)
}

type DataStore struct {
	DB *gorm.DB
}

type AuthResponse struct{
	Token string `json:"token"`
}

type User struct {
	UserID int `gorm:"Column:user_id;primary_key;auto_increment:true"`
	Username string `gorm:"Column:username;type:varchar(50)"`
	Password string `gorm:"type:varchar(400)"`
	Email string `gorm:"type:varchar(100);unique"`
	Token string `gorm:"type:varchar(300)"`
	Key string `gorm:"Column:public_key;type:varchar(500)"`
}

func (d *DataStore) SaveUser(user User) (AuthResponse, error) {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	generatedToken, err := token.SignedString([]byte("secret"))
	if err != nil {
		return AuthResponse{}, err
	}
	user.Token = generatedToken
	createdUser := d.DB.Create(&user)
	err = createdUser.Error
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{Token:generatedToken}, nil
}

func (d *DataStore) FetchUser(email string) (User, error) {
	user := User{}
	u := d.DB.Where("email = ?", email).First(&user)
	err := u.Error
	if err != nil {
		return User{}, err
	}
	fmt.Printf("%+v\n", user)
	return user, nil
}