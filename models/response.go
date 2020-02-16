package models

type Response struct {
	Data string `json:"data"`
	Errors []string `json:"errors"`
}