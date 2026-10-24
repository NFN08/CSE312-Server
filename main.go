package main

import "312WebAppNnero/utils"
import "net"
import "fmt"
import "github.com/joho/godotenv"
import "log"
import "os"
import "context"
import "github.com/jackc/pgx/v5/pgxpool"

func main() {
	route := utils.Route{}
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	conn, err := pgxpool.New(
		context.Background(),
		// "postgres://"+os.Getenv("LOCAL_DB_USER")+":"+os.Getenv("LOCAL_DB_PASSWORD")+"@postgres:5432/"+os.Getenv("LOCAL_DB"))
		"postgres://"+os.Getenv("LOCAL_DB_USER")+":"+os.Getenv("LOCAL_DB_PASSWORD")+"@localhost:5432/"+os.Getenv("LOCAL_DB"))

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
        userId TEXT,
		claimed BOOLEAN DEFAULT false,
		propic TEXT
    )`,
	)
	if err != nil {
		log.Fatal("error creating chat table:", err)
	}
	_, err = conn.Exec(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS videos (
    	userid TEXT,
        id TEXT,
		author TEXT,
		title TEXT,
		description TEXT,
		video_path TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		thumbnailURL TEXT,
		status TEXT,
		hls_path TEXT
		
    )`)
	if err != nil {
		log.Fatal("error creating users table", err)
	}

	_, err = conn.Exec(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS users (
        username TEXT UNIQUE NOT NULL,
    	userid TEXT,
        hash TEXT,
		authtoken TEXT,
		displayname TEXT,
		xsrftoken TEXT,
		github_id BIGINT UNIQUE,
		propic TEXT,
		videourl TEXT
    )`)
	if err != nil {
		log.Fatal("error creating users table", err)
	}

	_, err = conn.Exec(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS guest_sessions (
    	userid TEXT,
		xsrf_token TEXT,
		author TEXT,
		oauth_state TEXT
    )`)

	if err != nil {
		log.Fatal("error creating users table:", err)
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
	route.SetRoute("GET", "/video-call-room", router.GetFile, &router)
	route.SetRoute("GET", "/video-call", router.GetFile, &router)
	route.SetRoute("GET", "/view-video", router.GetFile, &router)
	route.SetRoute("POST", "/api/chats", router.ChatHandler, &router)
	route.SetRoute("GET", "/api/chats", router.GetMessages, &router)
	route.SetRoute("PATCH", "/api/chats/", router.UpdateMessage, &router)
	route.SetRoute("DELETE", "/api/chats/", router.DeleteMessage, &router)
	route.SetRoute("GET", "/public/cat.jpg", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/dog.jpg", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/eagle.jpg", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/elephant-small.jpg", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/elephant.jpg", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/favicon.ico", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/flamingo.jpg", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/jumping-cat.gif", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/kitten.jpg", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/user.webp", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/alertManager.js", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/chat.js", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/search-users.js", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/settings.js", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/utils.js", router.GetPublicFile, &router)
	route.SetRoute("GET", "/public/videos.js", router.GetPublicFile, &router)
	route.SetRoute("GET", "/login", router.GetFile, &router)
	route.SetRoute("GET", "/register", router.GetFile, &router)
	route.SetRoute("POST", "/register", router.Registration, &router)
	route.SetRoute("POST", "/login", router.Login, &router)
	route.SetRoute("POST", "/logout", router.Logout, &router)
	route.SetRoute("GET", "/api/users/@me", router.GetProfile, &router)
	route.SetRoute("POST", "/api/users/settings", router.UpdateCreds, &router)
	route.SetRoute("GET", "/authgithub", router.AuthGithub, &router)
	route.SetRoute("GET", "/authcallback", router.AuthCallback, &router)
	route.SetRoute("POST", "/api/users/avatar", router.ProPic, &router)
	route.SetRoute("POST", "/api/videos", router.UploadVideo, &router)
	route.SetRoute("GET", "/api/videos", router.GetVideos, &router)
	route.SetRoute("GET", "/api/videos/{videoID}", router.GetVideo, &router)
	route.SetRoute("GET", "/videotube", router.GetFile, &router)
	route.SetRoute("GET", "/videotube/upload", router.GetFile, &router)
	route.SetRoute("GET", "/videotube/videos/{videoID}", router.GetFile, &router)

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
		log.Fatal("error is", err)
	}
	for {
		conn, err := ln.Accept()

		if err != nil {
			// handle error
			fmt.Println("error is", err)
			continue
		}
		go handleConnection(conn, &router)
	}

}

func handleConnection(conn net.Conn, router *utils.Router) {
	defer conn.Close()

	for {
		requestObj := utils.Request{}
		response := utils.Response{}
		requestObj = utils.Request{
			Conn: conn,
		}
		//handle larger requests somehow
		bytes := make([]byte, 4096)
		n, err := conn.Read(bytes)
		if err != nil {
			fmt.Println("Encountered an error", err)
			return
		}
		// requestStr := string(bytes[:n])
		fmt.Println("request", string(bytes))
		utils.Parse(bytes[:n], &requestObj)
		closeConnection := utils.ParseConnection(requestObj.Headers)
		router.RouteTo(&requestObj, &response)
		conn.Write(response.Bytes)
		if closeConnection {
			return
		}

	}
}
