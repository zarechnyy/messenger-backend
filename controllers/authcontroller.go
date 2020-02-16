package controllers

import (
	"MyMessenger/logger"
	"MyMessenger/models"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
	"io/ioutil"
	"net/http"
	"os"
)

type AuthController struct {
	DataStore models.DataStorer
}

var logr logger.Logger

func (controller *AuthController) SignUpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "No POST", r.Method)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		requestBody := struct {
			Key      string
			Name     string
			Email    string
			Password string
		}{}

		if err := json.Unmarshal(body, &requestBody); err != nil {
			logr.LogErr(err)
			return
		}

		privateKey, err := getServerPrivateKey()
		if err != nil {
			println("getServerPrivateKey err")
			logr.LogErr(err)
			return
		}

		data, err := base64.StdEncoding.DecodeString(requestBody.Password)

		plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
		if err != nil {
			logr.LogErr(err)
			return
		}

		hashedPw := hashPassword(plaintext)
		user := models.User{
			Username: requestBody.Name,
			Password: hashedPw,
			Email:    requestBody.Email,
			Key:      requestBody.Key,
		}

		response, err := controller.DataStore.SaveUser(user)
		if err != nil {
			logr.LogErr(err)
			return
		}

		json.NewEncoder(w).Encode(response)
	}
}

func (controller *AuthController) LoginInHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "No POST", r.Method)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		requestBody := struct {
			Email    string
			Password string
		}{}

		if err := json.Unmarshal(body, &requestBody); err != nil {
			logr.LogErr(err)
			return
		}

		privateKey, err := getServerPrivateKey()
		if err != nil {
			println("getServerPrivateKey err")
			logr.LogErr(err)
			return
		}

		data, err := base64.StdEncoding.DecodeString(requestBody.Password)

		plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
		if err != nil {
			logr.LogErr(err)
			return
		}

		hashedPw := hashPassword(plaintext)
		res, err := controller.DataStore.FetchUser(requestBody.Email)

		if err != nil {
			logr.LogErr(err)
			return
		}

		if res.Password == hashedPw {
			json.NewEncoder(w).Encode(models.AuthResponse{Token: res.Token})
		} else {
			json.NewEncoder(w).Encode(errors.New("something went wrong"))
		}
	}
}

func getServerPrivateKey() (*rsa.PrivateKey, error) {
	return ParseRsaPrivateKeyFromPemStr(os.Getenv("SERVER_PRIVATE_KEY"))
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	mystr := "-----BEGIN RSA PRIVATE KEY-----\n" + privPEM + "\n-----END RSA PRIVATE KEY-----"
	block, _ := pem.Decode([]byte(mystr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		println("ParsePKCS1PrivateKey err")
		println(err.Error())
		return nil, err
	}

	return priv, nil
}

func hashPassword(pw []byte) string {
	salt := "mysalt"
	tempPassword := pbkdf2.Key(pw, []byte(salt), 10000, 50, sha256.New)
	return fmt.Sprintf("%x", tempPassword)
}
