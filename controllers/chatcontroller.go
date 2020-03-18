package controllers

import (
	"MyMessenger/models"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"path"

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
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logr.LogErr(err)
			return
		}
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			logr.LogErr(errors.New("No token"))
			return
		}
		//client A
		reqToken = splitToken[1]

		println(reqToken)
		//ID client B
		urlString := r.RequestURI
		idClientB := path.Base(urlString)

		println(idClientB)
		isValid, err := c.DataStore.IsValidToken(reqToken)

		if err != nil || isValid {
			println("SRAKA 1")
			return
		}

		selfUser, err := c.DataStore.FetchSelfUser(reqToken)

		if err != nil {
			println("SRAKA 2")
			return
		}
		println("SELF USER")
		println(selfUser.UserID)

		clientB, err := c.DataStore.FindUserBy(reqToken, idClientB)
		if err != nil {
			println("SRAKA 3")
			return
		}
		println("CLIENT B USER")
		println(clientB.UserID)
		newChat := models.ChatConnection {
			SelfUser: &selfUser,
			User:     &clientB,
			Ws:       ws,
		}

		//chats[newChat] = true
		selfClient := models.Client {
			User: &selfUser,
			Ws:   ws,
		}

		//подписать приватным ключем
		msg := models.SocketCommand {
			Type:    0,
			Message: clientB.Key,
		}

		clients[selfClient] = true
		selfClient.Ws.WriteJSON(msg)

		err = selfClient.Ws.ReadJSON(&msg)
		if err != nil {
			println("SRAKA 4")
			return
		}
		//if val, ok := clients[clientB.UserID]; !ok {
		//	println("SRAKA 5")
		//	return
		//}

		defer ws.Close()

		//go c.HandleMessages(ws, client, toID)
		//go c.HandleConnection(&newChat)
	}
}



//func (c *Controller) HandleConnection(chat *models.ChatConnection) {
//
//	chat.Ws.WriteJSON()
//	for {
//
//	}
//}