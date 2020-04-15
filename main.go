package main

import (
	"MyMessenger/controllers"
	"MyMessenger/driver"
	"MyMessenger/models"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	setUpKeys()

	db := driver.ConnectDB()
	storeData := models.DataStore{DB: db}

	controller := controllers.AuthController{DataStore: &storeData}
	userController := controllers.UsersController{DataStore: &storeData}
	chatController := controllers.Controller{DataStore: &storeData}
	router := mux.NewRouter()
	router.HandleFunc("/signup", controller.SignUpHandler()).Methods("POST")
	router.HandleFunc("/login", controller.LoginInHandler()).Methods("POST")
	router.HandleFunc("/users", userController.GetAllUsers()).Methods("GET")
	router.HandleFunc("/users/{text}", userController.FindUser()).Methods("GET")
	router.HandleFunc("/chat", chatController.HandleChatConnection()).Methods("POST")
	router.HandleFunc("/ws", chatController.HandleChatLogic())
	//router.HandleFunc("/keys", chatController.HandleChatLogic())

	fmt.Println("Server is listening...")
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe(":8181", loggedRouter))
}

func setUpKeys() {
	os.Setenv("SERVER_PRIVATE_KEY", "MIIEpQIBAAKCAQEApgFFeAcjR/t0rqBB/PGQ6SdD46oPF6E9juQZIe7I8H+EdxWjL6L3UBWnqsXyzFsYle+VakXBba8KOec3K4FLEq9OR1upeadY+QGbgZcaCxqR3jMpx2Z6psYbNG+CcnAQAx8DAt6rHPC+SUqm7VnZLqvg9NuEZZL/pHR89vIhGkWrgvtCcetJ/LdQPwSV4lLSGv3h1OQ8+05zOwGkcBFsIWZ5sgu7XZWZ1HYhb8v+LBVkg85+W7Dap7M3I3PQj5sYea/CWGEzR7r0TIs8K/oo0aqrqOQ2Wqms5YBIeEB4b2sdkxM/He+CN5TXKJWkCfb6j5eLCuzSdveiLBnV31HotQIDAQABAoIBAQCSGHpX6Qg+2GlXKgkJIFvlJb5UxZyksi3n7IzF1U2YgtFEsJE+YlD/dR9rZuyItv9LLWA0+BEEg9EfJoLiUdaNKiVdHaloPERMWQLPqPitnwOTJzn0mBcHYUAeBKksJ72f0ZIn22mCfckp3X5NUw6VSYUgWXUKo7VCuZYlCvDhGdSNKnSu2M+07gMNos6rFrcnnQxS7BD1Aizk1hGewosyfq6cq3UJx/0vXKex14Fi4ifYzPFNO7d1gmp75KKtChbVc8RVO6Ys/StFJxMuKYx5Pou1wqZdBjORoqgS5Rf0kA6QSvz5Mdb/By7ZwvMf6Nlgo8XmNtqJegqdCG48yF09AoGBAOYUa1XLdQp0Lx7FC65S+/U+fVtHwEAhoFHYhLgPqL1Hdjha1Ss++wtmAceiEzqp4oid8Iw0kjx0grXLoiowxLRLSbMasWbz90aVmiorvqxehwYEwxbWWHJWkKPmaQqU0crzO5UKMQAkEkwBxzl5SLmOZz6xDsVqIa9+A1xdN/tbAoGBALi06Z5XGoPX2XVCLkoioSz3U9fnqYsD9x4iu753JGoaYbvy07f3qiB3PlMyZ25xW0b6KSkUH2jNvS3wBzzpsoPx/OT39TFE+UW+2X3Q02nd29UGexLMK3sjr/b+rXEEAXeoY3jKhLjn3YHHFzi/eLllp1h5AS/1Is4JDOmzErkvAoGBAJor9pyf4AaoQebhbOlcK/9y5zciRj3zCmWtq4lW3OAwoZQzsYHwCvLhYLHv9eiqa+TVyJl6pL8j526ATGLvGPAjPvhoG5X8RqcimhJGC9ee4+VxjXShHtVHElbxj1OK02WmRTeig6Evip8p1eC6V7QXKzHEHTzF2FqrGv9qa5ffAoGAO466NcBIYHLdP54TZvw7lFA7zMZ6OMUSjbkNaKDqMPxIv13RPuSxCr7obdM23rnWgNBxLTm71wNgGMvoyY9hbII+1WXOvhBLgF3Fq3gGc4CCPfJVBP6olpAvUSlVq7dq8bZuPKiwmx7IoewcZMP4nW9VwoViCKC2lFD+xOxlASkCgYEAgXDs6kfqUNQvviRf5K2RSTe9tcToPa+bGWLuhgSHDK6PGA0NFGDLc4R5tGzYGc2XivFAXLp81B1Ym2q/ElnuRhVEB+ovhscWlkLDB8NZMxOsVJqUzuhs9Arf9kbMIQwJNgwp8Q9ZEZaIjWvr/LWbQ33eIAJt/FHMrhb7rc8yIUk=")
	os.Setenv("SERVER_PUBLIC_KEY", "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApgFFeAcjR/t0rqBB/PGQ6SdD46oPF6E9juQZIe7I8H+EdxWjL6L3UBWnqsXyzFsYle+VakXBba8KOec3K4FLEq9OR1upeadY+QGbgZcaCxqR3jMpx2Z6psYbNG+CcnAQAx8DAt6rHPC+SUqm7VnZLqvg9NuEZZL/pHR89vIhGkWrgvtCcetJ/LdQPwSV4lLSGv3h1OQ8+05zOwGkcBFsIWZ5sgu7XZWZ1HYhb8v+LBVkg85+W7Dap7M3I3PQj5sYea/CWGEzR7r0TIs8K/oo0aqrqOQ2Wqms5YBIeEB4b2sdkxM/He+CN5TXKJWkCfb6j5eLCuzSdveiLBnV31HotQIDAQAB")
}