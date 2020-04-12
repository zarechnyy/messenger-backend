package models

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	"time"
)

type DataStorer interface {
	SaveUser(user User) (AuthResponse, error)
	FetchUser(username string) (User, error)
	GetAllUsers(token string) ([]User, error)
	IsValidToken(token string) (bool, error)
	FindUser(token string, name string) ([]User, error)
	FetchSelfUser(token string) (User, error)
	FindUserBy(token string, id int) (User, error)
}

type DataStore struct {
	DB *gorm.DB
}

type AuthResponse struct {
	Token string `json:"token"`
	User User `json:"user"`
}

type User struct {
	UserID   int    `gorm:"Column:user_id;primary_key;auto_increment:true" json:"id"`
	Username string `gorm:"Column:username;type:varchar(50)" json:"username"`
	Password string `gorm:"type:varchar(400)" json:"-"`
	Email    string `gorm:"type:varchar(100);unique" json:"-"`
	Token    string `gorm:"type:varchar(300)" json:"-"`
	Key      string `gorm:"Column:public_key;type:varchar(500)" json:"-"`
}

type Room struct {
	ClientA Client
	ClientB Client
}

type Client struct {
	User *User
	Ws *websocket.Conn
}

type SocketCommand struct {
	Type int `json:"type"`
	Model interface{} `json:"model"`
}

type SocketKeyModel struct {
	Key string `json:"key"`
	Iv string `json:"iv"`
	SignatureKey string `json:"signatureKey"`
	SignatureIv string `json:"signatureIv"`
}

type SocketMessageModel struct {
	Message string `json:"message"`
}

type SocketDataModel struct {
	Data []byte `json:"data"`
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

	return AuthResponse{Token: generatedToken, User:user}, nil
}

func (d *DataStore) FetchUser(email string) (User, error) {
	user := User{}
	u := d.DB.Where("email = ?", email).First(&user)
	err := u.Error
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (d *DataStore) GetAllUsers(token string) ([]User, error) {
	var users []User
	req := d.DB.Where("token <> ?", token).Find(&users)
	err := req.Error
	if err != nil {
		return users, errors.New("empty kek")
	}
	return users, nil
}

func (d *DataStore) FindUser(token string, name string) ([]User, error) {
	var users []User
	searchPattern := "%" + name + "%"
	println(searchPattern)
	req := d.DB.Where("token <> ? AND username LIKE ?", token, searchPattern).Find(&users)
	err := req.Error
	if err != nil {
		return users, errors.New("empty kek")
	}
	return users, nil
}

func (d *DataStore) FindUserBy(token string, id int) (User, error) {
	var user User
	req := d.DB.Where("token <> ? AND user_id = ?", token, id).First(&user)
	err := req.Error
	if err != nil {
		return user, errors.New("empty kek")
	}
	return user, nil
}

func (d *DataStore) IsValidToken(token string) (bool, error) {
	var user User
	req := d.DB.Where("token = ?", token).First(&user)
	if req.Error != nil {
		return false, req.Error
	}
	if user.UserID == 0 {
		return false, errors.New("no valid token")
	}
	return true, nil
}

func (d *DataStore) FetchSelfUser(token string) (User, error) {
	var user User
	req := d.DB.Where("token = ?", token).First(&user)
	if req.Error != nil {
		return User{}, req.Error
	}
	if user.UserID == 0 {
		return User{}, errors.New("no valid token")
	}
	return user, nil
}