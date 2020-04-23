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
	"sync"
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
		var wg sync.WaitGroup

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
				delete(rooms, currentRoom)
				wg.Add(1)
				currentRoom  = createRoom(ws, &selfUser, &clientB)

				go keyTrade(ws, currentRoom, &wg)
			} else {
				currentRoom.ChatChannel <- ws
				flag := <- currentRoom.BoolChannel

				if !flag {
					delete(rooms, currentRoom)
					wg.Add(1)
					currentRoom  = createRoom(ws, &selfUser, &clientB)

					go keyTrade(ws, currentRoom, &wg)
				}
			}
		} else {
			wg.Add(1)
			currentRoom = createRoom(ws, &selfUser, &clientB)

			go keyTrade(ws, currentRoom, &wg)
		}

		wg.Wait()

		if currentRoom.ClientA.Ws == ws && !pingClient(ws, currentRoom) { return }

		println(fmt.Sprintf("CHATTING %s ", selfUser.Username))
		go chatting(ws, currentRoom)
	}
}

func createRoom(ws *websocket.Conn, selfUser *models.User, userB *models.User) *models.Room {

	room := models.Room{
		ClientA: models.Client{},
		ClientB: models.Client{},
	}

	room.ClientA.Ws = ws
	room.ClientA.User = selfUser
	room.ChatChannel = make(chan *websocket.Conn)
	room.BoolChannel = make(chan bool)

	room.ClientB.User = userB

	rooms[&room] = true
	return  &room
}

func keyTrade(ws *websocket.Conn, currentRoom *models.Room, wg *sync.WaitGroup) {
	defer wg.Done()

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
	fmt.Printf("%+v\n", currentRoom)

	currentRoom.ClientB.Ws = <- currentRoom.ChatChannel

	msg.Type = 2
	msg.Model = socketModel

	if err = currentRoom.ClientB.Ws.WriteJSON(msg); err != nil {
		handleWsError(currentRoom, ws)
		println(err.Error())
	}
}

func chatting(ws *websocket.Conn, room* models.Room ) {
	for {
		msg := models.SocketCommand{}
		err := ws.ReadJSON(&msg)
		if err != nil {
			logr.LogErr(err)
			handleWsError(room, ws)
			return
		}

		model := models.SocketDataModel{}
		if err := mapstructure.Decode(msg.Model, &model); err != nil {
			logr.LogErr(err)
			return
		}
		msg.Model = model

		if ws == room.ClientA.Ws {
			userName := room.ClientA.User.Username
			println(fmt.Sprintf("Client %s writes to %s", userName, room.ClientB.User.Username))
			err = room.ClientB.Ws.WriteJSON(msg)
		} else {
			userName := room.ClientB.User.Username
			println(fmt.Sprintf("Client %s writes to %s", userName, room.ClientA.User.Username))
			err = room.ClientA.Ws.WriteJSON(msg)
		}
		fmt.Printf("%+v\n", msg)

		if err != nil {
			logr.LogErr(err)
			handleWsError(room, ws)
			return
		}
	}
}

func pingClient(ws *websocket.Conn, currentRoom *models.Room) bool {
	pingMsg := models.SocketCommand{
		Type:  -1,
		Model: models.SocketMessageModel{Message:"ping"},
	}

	if err := ws.WriteJSON(pingMsg); err != nil {
		currentRoom.BoolChannel <- false
		ws.Close()
		return false
	}

	err := ws.ReadJSON(&pingMsg)
	if err != nil {
		currentRoom.BoolChannel <- false
		ws.Close()
		return false
	}

	currentRoom.BoolChannel <- true
	return true
}

func handleWsError(room *models.Room, closedWs *websocket.Conn) {

	msg := models.SocketCommand{}
	msg.Type = 5
	msg.Model = models.SocketMessageModel{Message:"CLOSE!"}

	if room.ClientA.Ws == closedWs && room.ClientB.Ws != nil {
		_ = room.ClientB.Ws.WriteJSON(msg)
		_ = room.ClientB.Ws.Close()
	} else if room.ClientA.Ws != nil {
		_ = room.ClientA.Ws.WriteJSON(msg)
		_ = room.ClientA.Ws.Close()
	}

	_ = closedWs.Close()
	fmt.Printf("%+v\n", room)
	delete(rooms, room)
}