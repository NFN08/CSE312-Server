package main

import "312WebAppNnero/utils"
import "net"
import "fmt"
import "github.com/joho/godotenv"
import "log"

import "os"
import "github.com/jackc/pgx/v5"
import "context"

//then upgrade to parse images,  maybe getting multiple requests check headers, double chck nosniff but it is showing up

func main() {
	route := utils.Route{}
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// use prepared statemetns even tho docs say unnessesary
	conn, err := pgx.Connect(
		context.Background(),
		"postgres://"+os.Getenv("LOCAL_DB_USER")+":"+os.Getenv("LOCAL_DB_PASSWORD")+"@postgres:5432/"+os.Getenv("LOCAL_DB"))
	if err != nil {
		log.Fatal("error connecting to postgres", err)
	}

	router := utils.Router{
		DB: conn,
	}

	_, err = conn.Exec(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS chat (
        author TEXT,
        id TEXT,
        content TEXT,
        updated BOOLEAN,
        userId TEXT
    )`,
	)
	if err != nil {
		fmt.Println("error creating chat table:", err)
	}

	route.SetRoute("GET", "/404", router.GetFile, &router)
	route.SetRoute("GET", "/index", router.GetFile, &router)
	route.SetRoute("GET", "/public", router.GetPublicFile, &router)
	route.SetRoute("GET", "/", router.GetFile, &router)
	route.SetRoute("GET", "/change-avatar", router.GetFile, &router)
	route.SetRoute("GET", "/chat", router.GetFile, &router)
	route.SetRoute("GET", "/direct-messaging", router.GetFile, &router)
	route.SetRoute("GET", "/drawing-board", router.GetFile, &router)
	route.SetRoute("GET", "/lecture", router.GetFile, &router)
	route.SetRoute("GET", "/login", router.GetFile, &router)
	route.SetRoute("GET", "/register", router.GetFile, &router)
	route.SetRoute("GET", "/search-users", router.GetFile, &router)
	route.SetRoute("GET", "/set-thumbnail", router.GetFile, &router)
	route.SetRoute("GET", "/settings", router.GetFile, &router)
	route.SetRoute("GET", "/test-websocket", router.GetFile, &router)
	route.SetRoute("GET", "/upload", router.GetFile, &router)
	route.SetRoute("GET", "/video-call-room", router.GetFile, &router)
	route.SetRoute("GET", "/video-call", router.GetFile, &router)
	route.SetRoute("GET", "/videotube", router.GetFile, &router)
	route.SetRoute("GET", "/view-video", router.GetFile, &router)
	route.SetRoute("POST", "/api/chats", router.ChatHandler, &router)
	route.SetRoute("GET", "/api/chats", router.GetMessages, &router)
	route.SetRoute("PATCH", "/api/chats/", router.UpdateMessage, &router)
	route.SetRoute("DELETE", "/api/chats/", router.DeleteMessage, &router)

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
		fmt.Println("error is", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println("error is", err)
		}
		go handleConnection(conn, &router)
	}

}

func handleConnection(conn net.Conn, router *utils.Router) {
	requestObj := utils.Request{}
	response := utils.Response{}
	//handle larger requests somehow
	bytes := make([]byte, 4096)
	n, err := conn.Read(bytes)
	if err != nil {
		fmt.Println("Encountered an error", err)
		conn.Close()
		return
	}
	requestStr := string(bytes[:n])
	utils.Parse(requestStr, &requestObj)
	router.RouteTo(&requestObj, &response)
	conn.Write(response.Bytes)
	defer conn.Close()
}
