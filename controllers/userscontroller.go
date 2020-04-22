package controllers

import (
	"MyMessenger/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
)

var usersOnline = make(map[*models.Client]bool) // online users

type UsersController struct {
	DataStore models.DataStorer
}

func (c *UsersController) GetAllUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "No GET", r.Method)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			logr.LogErr(errors.New("No token"))
			return
		}
		reqToken = splitToken[1]
		isValid, err := c.DataStore.IsValidToken(reqToken)
		if err != nil || !isValid {
			logr.LogErr(err)
			println(isValid)
			return
		}
		users, err := c.DataStore.GetAllUsers(reqToken)
		if err != nil {
			logr.LogErr(err)
			return
		}
		response := struct {
			Users []models.User `json:"users"`
		}{}
		response.Users = users
		json.NewEncoder(w).Encode(response)
	}
}

func (c *UsersController) ShowOnlineUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)

		defer ws.Close()

		if err != nil {
			logr.LogErr(err)
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

		client := models.Client{
			User: &selfUser,
			Ws:   ws,
		}
		usersOnline[&client] = true

		go sendUsers()

		msg := models.SocketCommand{}

		err = ws.ReadJSON(&msg)
		if err != nil {
			delete(usersOnline, &client)
			go sendUsers()
			return
		}

		if msg.Type == 6 {
			delete(usersOnline, &client)
			go sendUsers()
		}
	}
}

func sendUsers() {
	keys := make([]models.Client, 0, len(usersOnline))
	for k := range usersOnline {
		keys = append(keys, *k)
	}

	for client, _ := range usersOnline {
		msg := models.SocketCommand{}
		response := struct {
			Users []models.User `json:"users"`
		}{}

		response.Users = []models.User{}

		for _, user := range removeElement(keys, *client) {
			response.Users = append(response.Users, *user.User)
		}

		msg.Type = 6
		msg.Model = response

		err := client.Ws.WriteJSON(msg)
		if err != nil {
			logr.LogErr(err)
			return
		}
	}
}

func removeElement(arr []models.Client, client models.Client) []models.Client{
	var userIndex int
	newSlice :=  make([]models.Client, len(arr))
	copy(newSlice, arr)

	for i, v := range newSlice {
		if v == client {
			userIndex = i
			break
		}
	}
	return append(newSlice[:userIndex], newSlice[userIndex+1:]...)
}

func (c *UsersController) FindUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "No GET", r.Method)
			return
		}
		urlString := r.RequestURI
		searchName := path.Base(urlString)
		w.Header().Set("Content-Type", "application/json")
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			logr.LogErr(errors.New("No token"))
			return
		}
		reqToken = splitToken[1]
		isValid, err := c.DataStore.IsValidToken(reqToken)
		if err != nil || !isValid {
			logr.LogErr(err)
			println(isValid)
			return
		}
		users, err := c.DataStore.FindUser(reqToken, searchName)
		if err != nil {
			logr.LogErr(err)
			return
		}
		response := struct {
			Users []models.User `json:"users"`
		}{}
		response.Users = users
		json.NewEncoder(w).Encode(response)
	}
}
