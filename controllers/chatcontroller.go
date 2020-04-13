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
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var rooms = make(map[*models.Room]bool) // connected clients
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
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			logr.LogErr(errors.New("No token"))
		}
		reqToken = splitToken[1]
		isValid, err := c.DataStore.IsValidToken(reqToken)
		if err != nil || !isValid {
			return
		}
		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		requestBody := struct {
			ID    int  `json:"id"`
		}{}

		if err := json.Unmarshal(body, &requestBody); err != nil {
			logr.LogErr(err)
			return
		}

		clientB, err := c.DataStore.FindUserBy(reqToken, requestBody.ID)

		if err != nil {
			return
		}

		message := []byte(clientB.Key)
		hashed := sha256.Sum256(message)

		privateKey, err := getServerPrivateKey()
		if err != nil {
			logr.LogErr(err)
			return
		}
		signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
		if err != nil {
			logr.LogErr(err)
		}
		response := struct {
			Key string `json:"key"`
			Signature string `json:"signature"`
		}{}

		response.Key = clientB.Key
		base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
		base64.StdEncoding.Encode(base64Text, signature)
		response.Signature = string(base64Text)
		json.NewEncoder(w).Encode(response)
	}
}

func (c *Controller) HandleChatLogic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		println("SOCKET ENDPOINT")
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logr.LogErr(err)
			ws.Close()
			return
		}
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			logr.LogErr(err)
			return
		}
		reqToken = splitToken[1]
		isValid, err := c.DataStore.IsValidToken(reqToken)
		if err != nil || !isValid {
			logr.LogErr(err)
			return
		}

		selfUser, err := c.DataStore.FetchSelfUser(reqToken)
		if err != nil {
			logr.LogErr(err)
			return
		}

		msg := models.SocketCommand{}
		err = ws.ReadJSON(&msg)
		if err != nil {
			logr.LogErr(err)
			ws.Close()
			return
		}
		if msg.Type != 0 {
			logr.LogErr(err)
			ws.Close()
			return
		}

		model := models.SocketMessageModel{}
		if err := mapstructure.Decode(msg.Model, &model); err != nil {
			logr.LogErr(err)
			return
		}

		userID, err := strconv.Atoi(model.Message)
		if err != nil {
			logr.LogErr(err)
			return
		}

		clientB, err := c.DataStore.FindUserBy(reqToken, userID)
		if err != nil {
			logr.LogErr(err)
			return
		}

		var currentRoom *models.Room
		for room := range rooms {
			if room.ClientA.User.UserID == selfUser.UserID && room.ClientB.User.UserID == clientB.UserID || room.ClientB.User.UserID == selfUser.UserID && room.ClientA.User.UserID == clientB.UserID {
				currentRoom = room
				break
			}
		}
		if currentRoom != nil {
			if currentRoom.ClientA.User.UserID == selfUser.UserID {
				currentRoom.ClientA.Ws = ws
				keyTrade(ws, currentRoom)
			} else {
				currentRoom.ClientB.Ws = ws
			}
		} else {
			room := models.Room{
				ClientA: models.Client{},
				ClientB: models.Client{},
			}
			room.ClientA.Ws = ws
			room.ClientA.User = &selfUser
			room.ClientB.User = &clientB
			rooms[&room] = true

			currentRoom = &room
			keyTrade(ws, currentRoom)
		}
		println("LINE 174")
		for currentRoom.ClientB.Ws == nil || currentRoom.ClientA.Ws == nil {}
		println("LINE 175")
		go chatting(currentRoom.ClientA.Ws, currentRoom.ClientB.Ws, currentRoom)
		go chatting(currentRoom.ClientB.Ws, currentRoom.ClientA.Ws, currentRoom)
	}
}

func chatting(WsA *websocket.Conn, WsB *websocket.Conn, room* models.Room) {
	for {
		msg := models.SocketCommand{}
		err := WsA.ReadJSON(&msg)
		if err != nil {
			logr.LogErr(err)
			handleWsError(room, WsA)
			return
		}
		model := models.SocketDataModel{}
		if err := mapstructure.Decode(msg.Model, &model); err != nil {
			logr.LogErr(err)
			return
		}
		msg.Model = model
		fmt.Printf("%+v\n", model)
		err = WsB.WriteJSON(msg)
		if err != nil {
			logr.LogErr(err)
			handleWsError(room, WsA)
			return
		}
	}
}

func keyTrade(ws *websocket.Conn, currentRoom *models.Room) {
	socketModel := models.SocketKeyModel{}
	msg := models.SocketCommand{}
	msg.Type = 1
	msg.Model = models.SocketMessageModel{Message: "Create key pls"}
	err := ws.WriteJSON(msg)
	fmt.Printf("%+v\n", msg)
	if err != nil {
		handleWsError(currentRoom, ws)
		println(err.Error())
		return
	}
	msg = models.SocketCommand{}
	msg.Model = models.SocketKeyModel{}

	err = ws.ReadJSON(&socketModel)
	if err != nil {
		handleWsError(currentRoom, ws)
		println(err.Error())
		return
	}
	fmt.Printf("%+v\n", socketModel)
	for currentRoom.ClientB.Ws == nil { }
	msg.Type = 2
	msg.Model = socketModel
	fmt.Printf("%+v\n", msg)
	if err = currentRoom.ClientB.Ws.WriteJSON(msg); err != nil {
		handleWsError(currentRoom, ws)
		println(err.Error())
	}
}

func handleWsError(room *models.Room, closedWs *websocket.Conn) {
	println("handleWsError")
	closedWs.Close()
	if room.ClientA.Ws == closedWs {
		//room.ClientB.Ws.WriteJSON(msg)
		room.ClientB.Ws.Close()
	} else {
		//room.ClientA.Ws.WriteJSON(msg)
		room.ClientA.Ws.Close()
	}
	room.ClientA.Ws = nil
	room.ClientB.Ws = nil
	fmt.Printf("%+v\n", room)
	//room.ClientB.Ws.Close()
	//delete(rooms, room)
}