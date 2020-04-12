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
