package controllers

import (
	"MyMessenger/models"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"strings"
)

var chats = make(map[models.ChatConnection]bool) // connected clients
var clients = make(map[models.Client]bool)
type Controller struct {
	DataStore models.DataStorer
}

var upgrader = websocket.Upgrader {
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *Controller) HandleChatConnection() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		println("1")

		reqToken := r.Header.Get("Authorization")
		println(reqToken)
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			logr.LogErr(errors.New("No token"))
		}
		//client token
		reqToken = splitToken[1]
		println(splitToken)
		println("2")
		println("TOKET = ")
		println(reqToken)
		isValid, err := c.DataStore.IsValidToken(reqToken)
		if err != nil || !isValid {
			println("SRAKA 1")
			return
		}
		println("3")
		//user is authorized
		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		requestBody := struct {
			ID    int  `json:"id"`
		}{}

		println("4")
		if err := json.Unmarshal(body, &requestBody); err != nil {
			logr.LogErr(err)
			return
		}

		println("5")
		println(requestBody.ID)
		clientB, err := c.DataStore.FindUserBy(reqToken, requestBody.ID)

		if err != nil {
			return
		}
		println("6")

		message := []byte(clientB.Key)
		println("Client B KEY = ")
		println(clientB.Key)
		hashed := sha256.Sum256(message)

		privateKey, err := getServerPrivateKey()
		if err != nil {
			logr.LogErr(err)
			return
		}
		println("7")
		signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
		if err != nil {
			logr.LogErr(err)
		}
		println("SIGNATURE = ")
		println(string(signature))
		response := struct {
			Key string `json:"key"`
			Signature string `json:"signature"`
		}{}

		response.Key = clientB.Key
		base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
		base64.StdEncoding.Encode(base64Text, signature)
		response.Signature = string(base64Text)
		fmt.Printf("%+v\n", response)
		json.NewEncoder(w).Encode(response)
	}
}

//func (c *Controller) HandleChatConnection() http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		ws, err := upgrader.Upgrade(w, r, nil)
//		if err != nil {
//			logr.LogErr(err)
//			return
//		}
//		reqToken := r.Header.Get("Authorization")
//		splitToken := strings.Split(reqToken, "Bearer ")
//		if len(splitToken) != 2 {
//			logr.LogErr(errors.New("No token"))
//			return
//		}
//		//client A
//		reqToken = splitToken[1]
//
//		println(reqToken)
//		//ID client B
//		urlString := r.RequestURI
//		idClientB := path.Base(urlString)
//
//		println(idClientB)
//		isValid, err := c.DataStore.IsValidToken(reqToken)
//
//		if err != nil || isValid {
//			println("SRAKA 1")
//			return
//		}
//
//		selfUser, err := c.DataStore.FetchSelfUser(reqToken)
//
//		if err != nil {
//			println("SRAKA 2")
//			return
//		}
//		println("SELF USER")
//		println(selfUser.UserID)
//
//		clientB, err := c.DataStore.FindUserBy(reqToken, idClientB)
//		if err != nil {
//			println("SRAKA 3")
//			return
//		}
//		println("CLIENT B USER")
//		println(clientB.UserID)
//		newChat := models.ChatConnection {
//			SelfUser: &selfUser,
//			User:     &clientB,
//			Ws:       ws,
//		}
//
//		//chats[newChat] = true
//		selfClient := models.Client {
//			User: &selfUser,
//			Ws:   ws,
//		}
//
//		//подписать приватным ключем
//		msg := models.SocketCommand {
//			Type:    0,
//			Message: clientB.Key,
//		}
//
//		clients[selfClient] = true
//		selfClient.Ws.WriteJSON(msg)
//
//		err = selfClient.Ws.ReadJSON(&msg)
//		if err != nil {
//			println("SRAKA 4")
//			return
//		}
//		//if val, ok := clients[clientB.UserID]; !ok {
//		//	println("SRAKA 5")
//		//	return
//		//}
//
//		defer ws.Close()
//
//		//go c.HandleMessages(ws, client, toID)
//		//go c.HandleConnection(&newChat)
//	}
//}
