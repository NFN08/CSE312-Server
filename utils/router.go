package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"html"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)
import "strconv"
import "golang.org/x/crypto/bcrypt"
import (
	"crypto/rand"
	"encoding/hex"
)

type Route struct {
	method string
	path   string
	toCall func(Request, *Response)
}
type Router struct {
	route []Route
	DB    *pgxpool.Pool
}
type ChatMessage struct {
	Author  string `json:"author"`
	Id      string `json:"id"`
	Content string `json:"content"`
	Updated bool   `json:"updated"`
	Userid  string `json:"-"`
	Propic  string `json:"imageURL"`
}

func (route *Route) SetRoute(method string, path string, toCall func(Request, *Response), router *Router) {

	route.method = method
	route.path = path
	route.toCall = toCall
	router.AddRoute(*route)

}

func (router *Router) AddRoute(route Route) {
	router.route = append(router.route, route)
}

func (router *Router) RouteTo(request *Request, response *Response) {
	publicPath := ""
	chatPath := ""
	if len(request.path) >= 7 {
		publicPath = request.path[:7]
	}
	if strings.HasPrefix(request.path, "/api/chats/") {
		chatPath = "/api/chats/"
	}
	if publicPath != "/public" && !strings.HasPrefix(request.path, "/authcallback?") {
		parts := strings.Split(request.path, ".")
		path := parts[0]
		request.path = path
	}
	routePath := strings.SplitN(request.path, "?", 2)[0]
	isVideoID := strings.HasPrefix(routePath, "/api/videos/") && len(routePath) > len("/api/videos/")

	for _, route := range router.route {
		if route.method == request.method && route.path == routePath {
			route.toCall(*request, response)
			return

		} else if route.method == request.method && publicPath == route.path {
			route.toCall(*request, response)
			return

		} else if route.method == request.method && chatPath == route.path {
			route.toCall(*request, response)
			return
		} else if route.method == request.method && isVideoID && route.path == "/api/videos/{videoID}" {
			route.toCall(*request, response)
			return
		}
	}
	if publicPath == "/public" {
		router.GetPublicFile(*request, response)
	} else {
		router.GetFile(*request, response)
	}

}

// check for passwords and usernames that dpnt fit criteria
func DecodeCreds(request string) (string, string) {
	request = strings.TrimSpace(request)
	hexValues := map[string]int{"a": 10, "b": 11, "c": 12, "d": 13, "e": 14, "f": 15, "A": 10, "B": 11, "C": 12, "D": 13, "E": 14, "F": 15}
	parts := strings.Split(request, "=")
	password := ""

	if parts[2] != "" {
		password = parts[2]
	}

	// what if this is in password
	usernameParts := strings.Split(parts[1], "&password")
	username := ""

	if usernameParts[0] != "" {
		username = usernameParts[0]
	}
	hex := ""
	num2 := 0
	num3 := 0
	count := 0
	decodedUsername := ""

	for i := 0; i < len(username); i++ {
		if username[i] == '%' {
			hex = username[i+1 : i+3]
			num := 0
			if (hex[0] >= 'a' && hex[0] <= 'f') || (hex[0] >= 'A' && hex[0] <= 'F') {
				num = hexValues[string(hex[0])]
			} else if hex[0] >= '0' && hex[0] <= '9' {
				num = int(hex[0] - '0')
			}

			num = num * 16

			if (hex[1] >= 'a' && hex[1] <= 'f') || (hex[1] >= 'A' && hex[1] <= 'F') {
				num2 = hexValues[string(hex[1])]
				num = num + num2
			} else if hex[1] >= '0' && hex[1] <= '9' {
				num3 = int(hex[1] - '0')
				num = num + num3
			}

			ascii := byte(num)
			if (ascii >= 'a' && ascii <= 'z') || (ascii >= 'A' && ascii <= 'Z') || (ascii >= '0' && ascii <= '9') {
				decodedUsername += string(ascii)
			} else {
				return "400", ""
			}

			count++
			i = i + 2
			continue
		}

		if (username[i] >= 'a' && username[i] <= 'z') || (username[i] >= 'A' && username[i] <= 'Z') || (username[i] >= '0' && username[i] <= '9') {
			decodedUsername += string(username[i])
		} else {
			return "400", ""
		}

		count++
	}

	hex = ""
	num2 = 0
	num3 = 0
	count = 0
	decodedPassword := ""

	for i := 0; i < len(password); i++ {
		if password[i] == '%' {
			hex = password[i+1 : i+3]
			num := 0
			if (hex[0] >= 'a' && hex[0] <= 'f') || (hex[0] >= 'A' && hex[0] <= 'F') {
				num = hexValues[string(hex[0])]
			} else if hex[0] >= '0' && hex[0] <= '9' {
				num = int(hex[0] - '0')
			}

			num = num * 16

			if (hex[1] >= 'a' && hex[1] <= 'f') || (hex[1] >= 'A' && hex[1] <= 'F') {
				num2 = hexValues[string(hex[1])]
				num = num + num2
			} else if hex[1] >= '0' && hex[1] <= '9' {
				num3 = int(hex[1] - '0')
				num = num + num3
			}

			ascii := byte(num)

			if (ascii >= 'a' && ascii <= 'z') || (ascii >= 'A' && ascii <= 'Z') || (ascii >= '0' && ascii <= '9') ||
				ascii == '!' ||
				ascii == '@' ||
				ascii == '#' ||
				ascii == '$' ||
				ascii == '%' ||
				ascii == '^' ||
				ascii == '&' ||
				ascii == '(' ||
				ascii == ')' ||
				ascii == '-' ||
				ascii == '_' ||
				ascii == '=' {

				decodedPassword += string(ascii)
			} else {
				return "", "400"
			}

			count++
			i = i + 2
			continue
		}
		if (password[i] >= 'a' && password[i] <= 'z') || (password[i] >= 'A' && password[i] <= 'Z') || (password[i] >= '0' && password[i] <= '9') ||
			password[i] == '!' ||
			password[i] == '@' ||
			password[i] == '#' ||
			password[i] == '$' ||
			password[i] == '%' ||
			password[i] == '^' ||
			password[i] == '&' ||
			password[i] == '(' ||
			password[i] == ')' ||
			password[i] == '-' ||
			password[i] == '_' ||
			password[i] == '=' {

			decodedPassword += string(password[i])
		} else {
			return "", "400"
		}

		count++
	}

	return decodedUsername, decodedPassword
}

func ParseBody(request string) string {

	parts := strings.Split(request, ":")
	if parts[0] != "" {
		message := strings.Split(parts[1], "}")
		return message[0]
	} else {
		return ""
	}

}
func ParseHeaders(request []map[string]string) string {
	// id := ""
	for j := range request {
		for key, value := range request[j] {
			if strings.ToLower(key) == "cookie" {
				//this was panicking might not work for other headers
				// parts := strings.Split(value, "=")
				// id = parts[1]
				return value

			}
		}
	}
	return ""
}

func ParseXSRF(request []map[string]string) string {
	// id := ""
	for j := range request {
		for key, value := range request[j] {
			if strings.ToLower(key) == "x-xsrf-token" {
				//this was panicking might not work for other headers
				// parts := strings.Split(value, "=")
				// id = parts[1]
				return value

			}
		}
	}
	return ""
}

// changed this func double check it works elsewhere
func ParseContentType(request []map[string]string) (string, string) {
	for j := range request {
		for key, value := range request[j] {
			if strings.ToLower(key) == "content-type" {
				parts := strings.SplitN(value, ";", 2)
				if len(parts) < 2 {
					return value, ""
				}
				return parts[0], strings.TrimPrefix(strings.TrimSpace(parts[1]), "boundary=")

			}
		}
	}
	return "", ""
}

func ParseContentLength(request []map[string]string) int {
	for j := range request {
		for key, value := range request[j] {
			if strings.ToLower(key) == "content-length" {
				num, err := strconv.Atoi(value)
				if err != nil {
					fmt.Println("Error", err)
					return 0
				}
				return num

			}
		}
	}
	return 0
}

func ParseEtag(request []map[string]string) string {
	for j := range request {
		for key, value := range request[j] {
			if strings.ToLower(key) == "if-none-match" {
				return value
			}
		}
	}
	return ""
}

func ParseConnection(request []map[string]string) bool {
	close := false
	for j := range request {
		for key, value := range request[j] {
			if strings.ToLower(key) == "connection" && value == "close" {
				close = true
				return close
			}
		}
	}
	return false
}

func ParseRange(request []map[string]string) string {
	for j := range request {
		for key, value := range request[j] {
			if strings.ToLower(key) == "range" {
				return value
			}
		}
	}
	return ""
}

func (router *Router) UpdateCreds(request Request, response *Response) {
	header, _ := ParseContentType(request.Headers)
	if header != "application/x-www-form-urlencoded" {
		response.Setversion(request.version)
		response.Setcode("415")
		response.Setmessage("Unsupported Media Type")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	existingCookie := ParseHeaders(request.Headers)
	hashedGuest := sha256.Sum256([]byte(existingCookie))
	hashedGuestString := hex.EncodeToString(hashedGuest[:])

	GivenXSRF := ParseXSRF(request.Headers)

	var storedXSRF string
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT xsrftoken FROM users WHERE authtoken = $1`,
		hashedGuestString,
	)
	err := row.Scan(&storedXSRF)

	if err != nil || GivenXSRF != storedXSRF {
		response.Setversion(request.version)
		response.Setmessage("Forbidden")
		response.Setcode("403")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	cookievalue := ParseHeaders(request.Headers)
	hashedToken := sha256.Sum256([]byte(cookievalue))
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	row = router.DB.QueryRow(
		context.Background(),
		`SELECT username, userid
		FROM users
		WHERE authtoken = $1`,
		hashedTokenString,
	)
	var username string
	var userid string

	err = row.Scan(&username, &userid)
	if err != nil {
		fmt.Println("Error fetching creds", err)
		//error response
		response.Setversion(request.version)
		response.Setcode("401")
		response.Setmessage("Unauthorized")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	//authenticated
	newusername, newpassword := DecodeCreds(string(request.body))
	if newusername == "400" || newpassword == "400" {
		//400response
		fmt.Println("Error bad username/password", err)
		//error response
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Wrong format")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return

	}
	if newusername == "" && newpassword == "" {
		// do nothing
		response.Setversion(request.version)
		response.Setcode("200")
		response.Setmessage("OK")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return

	}
	if newusername != "" {
		//update username
		content := html.EscapeString(newusername)
		_, err = router.DB.Exec(
			context.Background(),
			`UPDATE users
			SET displayname = $1
    		WHERE userid = $2`,
			content, userid,
		)
		if err != nil {
			response.Setversion(request.version)
			response.Setcode("500")
			response.Setmessage("Internal Server Error")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		_, err = router.DB.Exec(
			context.Background(),
			`UPDATE chat
			SET author = $1
    		WHERE userid = $2`,
			content, userid,
		)
		if err != nil {
			//handle error
			fmt.Println("Error updating displayname", err)
			//error response
			response.Setversion(request.version)
			response.Setcode("400")
			response.Setmessage("Error")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return

		}

	}
	if newpassword != "" {
		if len(newpassword) < 8 {
			//return 400 resppnse
			response.Setversion(request.version)
			response.Setcode("400")
			response.Setmessage("Error password to short")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
		} else {
			hasLower := false
			hasUpper := false
			hasNumber := false
			hasSpecial := false

			for i := 0; i < len(newpassword); i++ {
				if newpassword[i] >= 'a' && newpassword[i] <= 'z' {
					hasLower = true
				} else if newpassword[i] >= 'A' && newpassword[i] <= 'Z' {
					hasUpper = true
				} else if newpassword[i] >= '0' && newpassword[i] <= '9' {
					hasNumber = true
				} else {
					hasSpecial = true
				}
			}

			if !hasLower || !hasUpper || !hasNumber || !hasSpecial {
				// return 400 response
				response.Setversion(request.version)
				response.Setcode("400")
				response.Setmessage("Error not enough entropy in password")
				response.Setnosniffheader()
				response.Setbodytext("")
				response.Setcontentlengthheader(0)
				response.Setallbytes()
			} else {
				// check db for dupe
				var usernameRet string
				row := router.DB.QueryRow(
					context.Background(),
					`SELECT username
					FROM users
					WHERE username = $1`,
					username,
				)
				err := row.Scan(&usernameRet)
				if err != nil {
					fmt.Println("Error fetching user", err)
					response.Setversion(request.version)
					response.Setcode("400")
					response.Setmessage("Error username already exists")
					response.Setnosniffheader()
					response.Setbodytext("")
					response.Setcontentlengthheader(0)
					response.Setallbytes()
					return
				}
				//update passowrd

				hash, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
				if err != nil {
					fmt.Println("Error salting password")
					//bad response
					response.Setversion(request.version)
					response.Setcode("400")
					response.Setmessage("Error")
					response.Setnosniffheader()
					response.Setbodytext("")
					response.Setcontentlengthheader(0)
					response.Setallbytes()
					return
				}
				_, err = router.DB.Exec(
					context.Background(),
					`UPDATE users
					SET hash = $1
    				WHERE userid = $2`,
					hash, userid,
				)
				if err != nil {
					response.Setversion(request.version)
					response.Setcode("500")
					response.Setmessage("Internal Server Error")
					response.Setnosniffheader()
					response.Setbodytext("")
					response.Setcontentlengthheader(0)
					response.Setallbytes()
					return
				}

				response.Setversion(request.version)
				response.Setcode("200")
				response.Setmessage("OK")
				response.Setnosniffheader()
				response.Setbodytext("")
				response.Setcontentlengthheader(0)
				response.Setallbytes()
				return
			}

		}
	}

	response.Setversion(request.version)
	response.Setcode("200")
	response.Setmessage("OK")
	response.Setnosniffheader()
	response.Setbodytext("")
	response.Setcontentlengthheader(0)
	response.Setallbytes()

}

// check registration/login with new changes
func (router *Router) Registration(request Request, response *Response) {
	header, _ := ParseContentType(request.Headers)
	if header != "application/x-www-form-urlencoded" {
		response.Setversion(request.version)
		response.Setcode("415")
		response.Setmessage("Unsupported Media Type")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	username, password := DecodeCreds(string(request.body))
	if username == "400" || password == "400" {
		//error response
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Wrong format")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	} else if username == "" || password == "" {
		//return 400 response
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Error invalid username or password")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	} else if strings.HasPrefix(strings.ToLower(username), "github") {
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Bad Request")
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	} else if len(password) < 8 {
		//return 400 resppnse
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Error password to short")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	} else {
		hasLower := false
		hasUpper := false
		hasNumber := false
		hasSpecial := false

		for i := 0; i < len(password); i++ {
			if password[i] >= 'a' && password[i] <= 'z' {
				hasLower = true
			} else if password[i] >= 'A' && password[i] <= 'Z' {
				hasUpper = true
			} else if password[i] >= '0' && password[i] <= '9' {
				hasNumber = true
			} else {
				hasSpecial = true
			}
		}

		if !hasLower || !hasUpper || !hasNumber || !hasSpecial {
			// return 400 response
			response.Setversion(request.version)
			response.Setcode("400")
			response.Setmessage("Error not enough entropy in password")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		} else {
			// check db for dupe
			var usernameRet string
			row := router.DB.QueryRow(
				context.Background(),
				`SELECT username
				FROM users
				WHERE username = $1`,
				username,
			)
			err := row.Scan(&usernameRet)
			if err == nil {
				fmt.Println("Error fetching user", err)
				response.Setversion(request.version)
				response.Setcode("400")
				response.Setmessage("Error username already exists")
				response.Setnosniffheader()
				response.Setbodytext("")
				response.Setcontentlengthheader(0)
				response.Setallbytes()
				return
			} else if err == pgx.ErrNoRows {
				// salt hash and store info user is legit
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					fmt.Println("Error salting password")
					//bad response
					response.Setversion(request.version)
					response.Setcode("400")
					response.Setmessage("Error")
					response.Setnosniffheader()
					response.Setbodytext("")
					response.Setcontentlengthheader(0)
					response.Setallbytes()
					return
				}
				id := string(gofakeit.ID())
				_, err = router.DB.Exec(
					context.Background(),
					`INSERT INTO users (username, userid, hash)
    				VALUES ($1, $2, $3)`,
					username,
					id,
					hash,
				)
				if err != nil {
					fmt.Println("Error adding user", err)
					//bad response
					response.Setversion(request.version)
					response.Setcode("400")
					response.Setmessage("Error adding user")
					response.Setnosniffheader()
					response.Setbodytext("")
					response.Setcontentlengthheader(0)
					response.Setallbytes()
					return
				}
				response.Setversion(request.version)
				response.Setcode("200")
				response.Setmessage("OK")
				response.Setnosniffheader()
				response.Setbodytext("")
				response.Setcontentlengthheader(0)
				response.Setallbytes()
			} else {
				// gen error
				fmt.Println("errr is", err)
				response.Setversion(request.version)
				response.Setcode("400")
				response.Setmessage("Error")
				response.Setnosniffheader()
				response.Setbodytext("")
				response.Setcontentlengthheader(0)
				response.Setallbytes()
				return
			}
		}

	}
}

// Converts the mp4 at inputPath into an HLS video with two
// qualities (720p and 360p) and saves the index file at outputDir/master.m3u8
func MP4ToHLS(inputPath string, outputDir string) error {
	return exec.Command("ffmpeg", "-i", inputPath,
		"-map", "0:v:0", "-map", "0:a:0", "-map", "0:v:0", "-map", "0:a:0",
		"-filter:v:0", "scale=-2:720", "-b:v:0", "2500k",
		"-filter:v:1", "scale=-2:360", "-b:v:1", "800k",
		"-f", "hls", "-hls_playlist_type", "vod",
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "v:0,a:0 v:1,a:1",
		outputDir+"/%v/index.m3u8",
	).Run()
}

func (router *Router) UploadVideo(request Request, response *Response) {
	authToken := ParseHeaders(request.Headers)
	hashedToken := sha256.Sum256([]byte(authToken))
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	var username string
	var userid string
	var authtoken string
	var storedXSRF string
	var author string
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT username, userid, authtoken
		 FROM users
		 WHERE authtoken = $1`,
		hashedTokenString,
	)
	err := row.Scan(&username, &userid, &authtoken)
	if err == nil {

		GivenXSRF := ParseXSRF(request.Headers)

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrftoken
			FROM users
			WHERE userid = $1`,
			userid,
		)

		err = row.Scan(&storedXSRF)

		if err != nil || GivenXSRF != storedXSRF {
			response.Setversion(request.version)
			response.Setmessage("Forbidden")
			response.Setcode("403")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		//logged in and valid XSRF token
		length := ParseContentLength(request.Headers)
		mimetype, boundary := ParseContentType(request.Headers)
		buffer := make([]byte, 4096)
		for len(request.body) < length {
			remaining := length - len(request.body)
			n, err := request.Conn.Read(buffer[:min(remaining, len(buffer))])
			if err != nil {
				fmt.Println("Encountered an error", err)
				return
			}
			request.body = append(request.body, buffer[:n]...)

		}

		if mimetype != "multipart/form-data" {
			//error response
			response.Setversion(request.version)
			response.Setmessage("Incorrect MimeType")
			response.Setcode("400")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return

		}
		var MPObj MultipartParts
		var err error
		// check ext matches approved ones
		MPObj = MultiPartParseVideo(request.body, boundary)
		if MPObj.Ext == "Wrong Ext" {
			response.Setversion(request.version)
			response.Setmessage("Invalid File Type")
			response.Setcode("400")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		if MPObj.Error != "error" {
			err = router.DB.QueryRow(
				context.Background(),
				`UPDATE users
        		 SET videourl = $1
         		WHERE userid = $2
				RETURNING username`,
				MPObj.Path,
				userid,
			).Scan(&author)
			if err != nil {
				fmt.Println("error adding path to db", err)
				return
			}

		} else {
			response.Setversion(request.version)
			response.Setmessage("Error Parsing")
			response.Setcode("400")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		status := "processing"
		thumbnailURL := "/public/imgs/avatars/defaultvidimg.jpg"
		_, err = router.DB.Exec(
			context.Background(),
			`INSERT INTO videos (userid, id, author, title, description, video_path, thumbnailURL, status)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
			userid, MPObj.Id, author, MPObj.Title, MPObj.Description, MPObj.Path, thumbnailURL,status,
		)
		if err != nil {
			fmt.Println("error inserting,", err)
			response.Setversion(request.version)
			response.Setmessage("Error")
			response.Setcode("401")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

		go func() {
			defer func() { recover() }()
			err := MP4ToHLS(strings.TrimPrefix(MPObj.Path, "/"), "public/videos/"+MPObj.Id)
			if err != nil {
				fmt.Println("hls conversion failed", err)
				router.DB.Exec(context.Background(), `UPDATE videos SET status = 'failed' WHERE id = $1`, MPObj.Id)
				return
			}
			hls_path := "/public/videos/"+MPObj.Id+"/master.m3u8"

			_, err = router.DB.Exec(context.Background(),
				`UPDATE videos SET status = 'ready', hls_path = $1 WHERE id = $2`,
				hls_path, MPObj.Id)
			if err != nil {
				fmt.Println("couldnt update hls info", err)
				return
			}
		}()

		body := fmt.Sprintf(`{"id":"%s"}`, MPObj.Id)
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		response.Setnosniffheader()
		response.Setbodyjson(body)
		response.Setcontentlengthheader(len(body))
		response.Setallbytes()
		return

	} else {
		//user isnt logged in
		response.Setversion(request.version)
		response.Setmessage("Unauthorized")
		response.Setcode("401")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return

	}

}

type MultipartParts struct {
	Headers     map[string]string
	Name        string
	Id          string
	Title       string
	Ext         string
	Description string
	Path        string
	Content     []byte
	Error       string
}

// store everything in file struct
func MultiPartParse(requestBody []byte, boundary string) string {
	CRLF := "\r\n"
	// this gives usthree sections hopefully
	_, part2, _ := bytes.Cut(requestBody, []byte(boundary))
	_, ext, found := bytes.Cut(part2, []byte("Content-Type:"))
	if !found {
		return "Wrong Ext"
	}
	mime, _, _ := bytes.Cut(ext, []byte("\r\n"))
	extension := strings.TrimSpace(string(mime))
	if extension == "image/jpeg" {
		extension = ".jpeg"
	} else if extension == "image/jpg" {
		extension = ".jpg"

	} else if extension == "image/png" {
		extension = ".png"
	} else if extension == "image/gif" {
		extension = ".gif"

	} else {
		return "Wrong Ext"
	}
	// changed first cut to ext like video should still work
	_, body1, _ := bytes.Cut(ext, []byte(CRLF+CRLF))
	body2, _, found := bytes.Cut(body1, []byte("\r\n--"+boundary))
	id := gofakeit.ID()

	err := os.WriteFile("public/imgs/avatars/"+id+string(extension), body2, 0644)
	if err != nil {
		fmt.Println("error", err)
		return "error"
	}

	return "/public/imgs/avatars/" + id + string(extension)
}

func MultiPartParseVideo(requestBody []byte, boundary string) MultipartParts {
	var MPObj MultipartParts
	CRLF := "\r\n"
	// this gives usthree sections hopefully
	_, part2, _ := bytes.Cut(requestBody, []byte(boundary))
	_, ext, found := bytes.Cut(part2, []byte("Content-Type:"))
	if !found {
		return MPObj
	}
	mime, _, _ := bytes.Cut(ext, []byte("\r\n"))
	extension := strings.TrimSpace(string(mime))
	if extension == "video/mp4" {
		extension = ".mp4"
		MPObj.Ext = extension
	}
	_, titlepart, _ := bytes.Cut(part2, []byte(CRLF+CRLF))
	title, _, _ := bytes.Cut(titlepart, []byte(CRLF))
	MPObj.Title = string(title)
	_, descriptionpart, _ := bytes.Cut(titlepart, []byte(CRLF+CRLF))
	description, _, _ := bytes.Cut(descriptionpart, []byte(CRLF))
	MPObj.Description = string(description)
	_, body1, _ := bytes.Cut(ext, []byte(CRLF+CRLF))
	body2, _, found := bytes.Cut(body1, []byte("\r\n--"+boundary))
	id := gofakeit.ID()
	MPObj.Id = id

	err := os.WriteFile("public/videos/"+id+string(extension), body2, 0644)
	if err != nil {
		fmt.Println("error", err)
		MPObj.Error = "error"
		return MPObj
	}
	path := "/public/videos/" + id + string(extension)
	MPObj.Path = path
	return MPObj
}

// make sure we set update old messages to include new pro pic
func (router *Router) ProPic(request Request, response *Response) {
	//check logged in
	authToken := ParseHeaders(request.Headers)
	hashedToken := sha256.Sum256([]byte(authToken))
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	var username string
	var userid string
	var authtoken string
	var storedXSRF string
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT username, userid, authtoken
		 FROM users
		 WHERE authtoken = $1`,
		hashedTokenString,
	)
	err := row.Scan(&username, &userid, &authtoken)
	if err == nil {

		GivenXSRF := ParseXSRF(request.Headers)

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrftoken
			FROM users
			WHERE userid = $1`,
			userid,
		)

		err = row.Scan(&storedXSRF)

		if err != nil || GivenXSRF != storedXSRF {
			response.Setversion(request.version)
			response.Setmessage("Forbidden")
			response.Setcode("403")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		//logged in and valid XSRF token
		length := ParseContentLength(request.Headers)
		mimetype, boundary := ParseContentType(request.Headers)
		buffer := make([]byte, 4096)
		for len(request.body) < length {
			remaining := length - len(request.body)
			n, err := request.Conn.Read(buffer[:min(remaining, len(buffer))])
			if err != nil {
				fmt.Println("Encountered an error", err)
				return
			}
			request.body = append(request.body, buffer[:n]...)

		}

		if mimetype != "multipart/form-data" {
			//error response
			response.Setversion(request.version)
			response.Setmessage("Incorrect MimeType")
			response.Setcode("400")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return

		}

		// check ext matches approved ones
		path := MultiPartParse(request.body, boundary)
		if path == "Wrong Ext" {
			response.Setversion(request.version)
			response.Setmessage("Invalid File Type")
			response.Setcode("400")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		if path != "error" {
			_, err = router.DB.Exec(
				context.Background(),
				`UPDATE users
        		 SET propic = $1
         		WHERE userid = $2`,
				path,
				userid,
			)
		} else {
			response.Setversion(request.version)
			response.Setmessage("Error Parsing")
			response.Setcode("400")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		if err != nil {
			fmt.Println("error adding path to db")
			return
		}
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return

	} else {
		//user isnt logged in
		response.Setversion(request.version)
		response.Setmessage("Unauthorized")
		response.Setcode("401")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return

	}

}

func (router *Router) AuthCallback(request Request, response *Response) {
	parts := strings.SplitN(request.path, "?", 2)

	if len(parts) != 2 {
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Bad Request")
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	var code string
	var state string

	parameters := strings.Split(parts[1], "&")

	for _, parameter := range parameters {
		pair := strings.SplitN(parameter, "=", 2)

		if len(pair) != 2 {
			continue
		}

		if pair[0] == "code" {
			code = pair[1]
		} else if pair[0] == "state" {
			state = pair[1]
		}
	}

	cookieValue := ParseHeaders(request.Headers)
	hashedCookie := sha256.Sum256([]byte(cookieValue))
	hashedGuestString := hex.EncodeToString(hashedCookie[:])
	if code == "" || state == "" || cookieValue == "" {
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Bad Request")
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	row := router.DB.QueryRow(
		context.Background(),
		`UPDATE guest_sessions
         SET oauth_state=NULL
         WHERE userid=$1 AND oauth_state=$2
         RETURNING userid`,
		hashedGuestString,
		state,
	)

	var sessionID string
	err := row.Scan(&sessionID)

	if err != nil {
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Bad Request")
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	data := strings.NewReader("client_id=" + os.Getenv("GITHUB_CLIENT_ID") + "&client_secret=" + os.Getenv("GITHUB_CLIENT_SECRET") + "&code=" + code)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", data)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	githubResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	var tokenData struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(githubResponse.Body).Decode(&tokenData)
	if err != nil || githubResponse.StatusCode != 200 || tokenData.AccessToken == "" {
		return
	}

	githubResponse.Body.Close()

	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

	githubResponse, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer githubResponse.Body.Close()

	var githubUser struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	}

	err = json.NewDecoder(githubResponse.Body).Decode(&githubUser)
	if err != nil || githubResponse.StatusCode != 200 || githubUser.ID == 0 {
		return
	}

	var userid string
	var username string
	var displayname *string

	row = router.DB.QueryRow(
		context.Background(),
		`SELECT userid, username, displayname
     	FROM users WHERE github_id=$1`,
		githubUser.ID,
	)

	err = row.Scan(&userid, &username, &displayname)

	if err == pgx.ErrNoRows {
		userid = gofakeit.ID()
		username = fmt.Sprintf("github%d", githubUser.ID)

		row = router.DB.QueryRow(
			context.Background(),
			`INSERT INTO users (userid, username, github_id)
         	VALUES ($1, $2, $3)
         	RETURNING userid`,
			userid, username, githubUser.ID,
		)

		err = row.Scan(&userid)
	}

	if err != nil {
		fmt.Println("error inserting", err)
		response.Setversion(request.version)
		response.Setcode("500")
		response.Setmessage("Internal Server Error")
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	randomBytes := make([]byte, 16)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return
	}

	authToken := hex.EncodeToString(randomBytes)
	hashedToken := sha256.Sum256([]byte(authToken))
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	XSRF := make([]byte, 16)
	_, err = rand.Read(XSRF)
	if err != nil {
		return
	}

	XSRFString := hex.EncodeToString(XSRF)

	_, err = router.DB.Exec(
		context.Background(),
		`UPDATE users
     	SET authtoken=$1, xsrftoken=$2
     	WHERE userid=$3`,
		hashedTokenString, XSRFString, userid,
	)

	if err != nil {
		return
	}
	if cookieValue != "" {
		author := username
		if displayname != nil {
			author = *displayname
		}

		_, err = router.DB.Exec(
			context.Background(),
			`UPDATE chat
         	SET author=$1, userid=$2, claimed=true
         	WHERE userid=$3 AND claimed=false`,
			author, userid, hashedGuestString,
		)

		if err != nil {
			return
		}
	}
	cookie := fmt.Sprintf("%s; Path=/", authToken)

	response.Setversion(request.version)
	response.Setcode("302")
	response.Setmessage("Found")
	response.Setheaders([]map[string]string{{"Set-Cookie": cookie + "; Max-Age=3600; HttpOnly; SameSite=Lax"}, {"Location": "/"}})
	response.Setnosniffheader()
	response.Setbodytext("")
	response.Setcontentlengthheader(0)
	response.Setallbytes()

}

func (router *Router) AuthGithub(request Request, response *Response) {
	cookieValue := ParseHeaders(request.Headers)
	hashedCookie := sha256.Sum256([]byte(cookieValue))
	hashedGuestString := hex.EncodeToString(hashedCookie[:])
	stateBytes := make([]byte, 16)
	_, err := rand.Read(stateBytes)
	if err != nil {
		return
	}

	state := hex.EncodeToString(stateBytes)
	_, err = router.DB.Exec(
		context.Background(),
		"UPDATE guest_sessions SET oauth_state=$1 WHERE userid=$2",
		state, hashedGuestString,
	)

	if err != nil {
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Failed to update")
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	authorizationURL := "https://github.com/login/oauth/authorize" + "?client_id=" + os.Getenv("GITHUB_CLIENT_ID") + "&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fauthcallback" + "&state=" + state
	response.Setversion(request.version)
	response.Setcode("302")
	response.Setmessage("Found")
	response.Setheaders([]map[string]string{{"Location": authorizationURL}})
	response.Setbodytext("")
	response.Setcontentlengthheader(0)
	response.Setallbytes()

}

// check double login only messages claimed once
func (router *Router) Login(request Request, response *Response) {
	header, _ := ParseContentType(request.Headers)
	if header != "application/x-www-form-urlencoded" {
		response.Setversion(request.version)
		response.Setcode("415")
		response.Setmessage("Unsupported Media Type")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	existingCookie := ParseHeaders(request.Headers)
	hashedGuest := sha256.Sum256([]byte(existingCookie))
	hashedGuestString := hex.EncodeToString(hashedGuest[:])

	GivenXSRF := ParseXSRF(request.Headers)

	var storedXSRF string
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT xsrf_token FROM guest_sessions WHERE userid = $1`,
		hashedGuestString,
	)

	err := row.Scan(&storedXSRF)

	if err != nil || GivenXSRF != storedXSRF {
		fmt.Println("error is", err)
		response.Setversion(request.version)
		response.Setmessage("Forbidden")
		response.Setcode("403")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	username, password := DecodeCreds(string(request.body))
	if username == "400" || password == "400" {
		//error response
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Wrong format")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	cookieVal := ParseHeaders(request.Headers)
	row = router.DB.QueryRow(
		context.Background(),
		`SELECT username, userid, hash, displayname
		FROM users
		WHERE username = $1`,
		username,
	)
	var userid string
	var hash string
	var displayname *string

	err = row.Scan(&username, &userid, &hash, &displayname)
	if err != nil {
		fmt.Println("Error getting info", err)
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Bad Request")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	err = bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
	if err != nil {
		//400 response not verified
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Error not verified")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	//authenticated
	randomBytes := make([]byte, 10)
	_, err = rand.Read(randomBytes)
	if err != nil {
		// handle error
		fmt.Println("couldnt gnereate auth token")
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Error coudlnt gen auth token")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	authToken := hex.EncodeToString(randomBytes)
	hashedToken := sha256.Sum256([]byte(authToken))
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	XSRF := make([]byte, 16)
	_, err = rand.Read(XSRF)
	if err != nil {
		// handle error
		fmt.Println("couldnt gnereate XSRF token")
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Error coudlnt gen XSRF token")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	XSRFString := hex.EncodeToString(XSRF)
	_, err = router.DB.Exec(
		context.Background(),
		`UPDATE users 
		 SET authtoken = $1, xsrftoken = $3
		 WHERE username = $2`,
		hashedTokenString, username, XSRFString,
	)
	if err != nil {
		fmt.Println("couldnt add hashed authtoken to db", err)
		//400 response
		response.Setversion(request.version)
		response.Setcode("400")
		response.Setmessage("Error coudlnt add authtoken to db")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	if cookieVal != "" {

		claimed := true
		hashedGuest := sha256.Sum256([]byte(cookieVal))
		hashedGuestString := hex.EncodeToString(hashedGuest[:])

		if displayname == nil {
			_, err = router.DB.Exec(
				context.Background(),
				`UPDATE chat
				SET author = $1, userid = $2, claimed = $3
    			WHERE userid = $4 AND claimed = false`,
				username, userid, claimed, hashedGuestString,
			)
		} else {
			_, err = router.DB.Exec(
				context.Background(),
				`UPDATE chat
				SET author = $1, userid = $2, claimed = $3
    			WHERE userid = $4 AND claimed = false`,
				displayname, userid, claimed, hashedGuestString,
			)
		}

	}

	cookie := fmt.Sprintf("%s; Path=/", authToken)
	text := "Success!"
	response.Setversion(request.version)
	response.Setmessage("OK")
	response.Setcode("200")
	response.Setheaders([]map[string]string{{"Set-Cookie": cookie + ";Max-Age=3600; HttpOnly"}})
	response.Setnosniffheader()
	response.Setbodytext(text)
	response.Setcontentlengthheader(len(text))
	response.Setallbytes()

}

// inconsistencies with logout not clearing auth token occasianlly
func (router *Router) Logout(request Request, response *Response) {
	cookievalue := ParseHeaders(request.Headers)
	hashedToken := sha256.Sum256([]byte(cookievalue))
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	GivenXSRF := ParseXSRF(request.Headers)

	var storedXSRF string
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT xsrftoken FROM users WHERE authtoken = $1`,
		hashedTokenString,
	)
	err := row.Scan(&storedXSRF)

	if err != nil || GivenXSRF != storedXSRF {
		response.Setversion(request.version)
		response.Setmessage("Forbidden")
		response.Setcode("403")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	_, err = router.DB.Exec(
		context.Background(),
		`UPDATE users
		 SET authtoken = NULL, xsrftoken = NULL
		 WHERE authtoken = $1`,
		hashedTokenString,
	)
	if err != nil {
		fmt.Println("Error setting authtoken null", err)
		return
	}

	response.Setversion(request.version)
	response.Setmessage("Found")
	response.Setcode("302")
	response.Setheaders([]map[string]string{{"Set-Cookie": cookievalue + ";Max-Age=0; HttpOnly"}, {"Location": "/"}})
	response.Setnosniffheader()
	response.Setbodytext("")
	response.Setcontentlengthheader(0)
	response.Setallbytes()

}

func (router *Router) GetVideos(request Request, response *Response) {

	rows, err := router.DB.Query(
		context.Background(),
		`SELECT id, author, title, description, video_path, created_at, thumbnailURL
     FROM videos`,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	videos := []map[string]string{}

	for rows.Next() {
		var id, author, title, description, videoPath, thumbnailURL string
		var createdAt time.Time

		err := rows.Scan(&id, &author, &title, &description, &videoPath, &createdAt, &thumbnailURL)
		if err != nil {
			fmt.Println(err)
			return
		}

		videos = append(videos, map[string]string{
			"id":           id,
			"author_id":    author,
			"title":        html.EscapeString(title),
			"description":  html.EscapeString(description),
			"video_path":   videoPath,
			"created_at":   createdAt.String(),
			"thumbnailURL": thumbnailURL,
		})
	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return
	}

	data, err := json.Marshal(map[string]interface{}{
		"videos": videos,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	response.Setversion(request.version)
	response.Setcode("200")
	response.Setmessage("Ok")
	response.Setnosniffheader()
	response.Setbodyjson(string(data))
	response.Setcontentlengthheader(len(data))
	response.Setallbytes()

}

func (router *Router) GetVideo(request Request, response *Response) {
	parts := strings.Split(request.path, "/")
	var videoid string
	if parts[3] != "" {
		videoid = parts[3]

	}

	rows := router.DB.QueryRow(
		context.Background(),
		`SELECT id, author, title, description, video_path, created_at, thumbnailURL, status, hls_path
     FROM videos
	 WHERE id = $1`,
		videoid,
	)

	video := map[string]string{}

	var id, author, title, description, videoPath, thumbnailURL, status string
	var createdAt time.Time
	var hls_path *string
	err := rows.Scan(&id, &author, &title, &description, &videoPath, &createdAt, &thumbnailURL, &status, &hls_path)
	if err == pgx.ErrNoRows {
		fmt.Println(err)
		response.Setversion(request.version)
		response.Setcode("404")
		response.Setmessage("No Video")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	hls := ""
	if hls_path != nil {
		hls = *hls_path
	}
	video = map[string]string{
		"id":           id,
		"author_id":    author,
		"title":        html.EscapeString(title),
		"description":  html.EscapeString(description),
		"video_path":   videoPath,
		"created_at":   createdAt.String(),
		"thumbnailURL": thumbnailURL,
		"status":       status,
		"hls_path":     hls,
	}

	data, err := json.Marshal(map[string]interface{}{
		"video": video,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	response.Setversion(request.version)
	response.Setcode("200")
	response.Setmessage("Ok")
	response.Setnosniffheader()
	response.Setbodyjson(string(data))
	response.Setcontentlengthheader(len(data))
	response.Setallbytes()

}

func (router *Router) GetProfile(request Request, response *Response) {
	cookievalue := ParseHeaders(request.Headers)
	hashedToken := sha256.Sum256([]byte(cookievalue))
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	row := router.DB.QueryRow(
		context.Background(),
		`SELECT username, userid, propic
		FROM users
		WHERE authtoken = $1`,
		hashedTokenString,
	)
	var username string
	var userid string
	var propic *string

	err := row.Scan(&username, &userid, &propic)
	if err != nil {
		fmt.Println("Error fetching profile not logged in", err)
		//error response
		response.Setversion(request.version)
		response.Setcode("401")
		response.Setmessage("Error coudlnt find info")
		response.Setnosniffheader()
		response.Setbodyjson("{}")
		response.Setcontentlengthheader(2)
		response.Setallbytes()
		return
	}
	imageURL := ""
	if propic != nil {
		imageURL = *propic
	}
	userData := map[string]string{
		"username": username,
		"id":       userid,
		"imageURL": imageURL,
	}
	data, err := json.Marshal(userData)
	if err != nil {
		fmt.Println("Error turning data into json")
		return
	}
	fullData := string(data)
	response.Setversion(request.version)
	response.Setmessage("Found")
	response.Setcode("200")
	response.Setnosniffheader()
	response.Setbodyjson(fullData)
	response.Setcontentlengthheader(len(fullData))
	response.Setallbytes()

}

// check that the guest cookie has 80 bits of entropy and hashed now
func (router *Router) DeleteMessage(request Request, response *Response) {
	parts := strings.Split(request.path, "/")
	messageid := parts[3]
	// why hashing twice here?
	authToken := ParseHeaders(request.Headers)
	hashedGuest := sha256.Sum256([]byte(authToken))
	hashedGuestString := hex.EncodeToString(hashedGuest[:])

	var author string
	var id string
	var content string
	var updated bool
	var userId string

	var username string
	var userid string
	var authtoken string

	hashedToken := sha256.Sum256([]byte(authToken))
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT username, userid, authtoken
		 FROM users
		 WHERE authtoken = $1`,
		hashedTokenString,
	)
	err := row.Scan(&username, &userid, &authtoken)
	if err == nil {

		GivenXSRF := ParseXSRF(request.Headers)

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrftoken
			FROM users
			WHERE userid = $1`,
			userid,
		)

		var storedXSRF string
		err = row.Scan(&storedXSRF)

		if err != nil || GivenXSRF != storedXSRF {
			response.Setversion(request.version)
			response.Setmessage("Forbidden")
			response.Setcode("403")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

		row = router.DB.QueryRow(
			context.Background(),
			`SELECT author, id, content, updated, userid FROM chat WHERE id = $1 AND userid = $2`,
			messageid, userid,
		)
		err = row.Scan(&author, &id, &content, &updated, &userId)
		updated = true
		if err == nil {

			_, err = router.DB.Exec(
				context.Background(),
				`DELETE FROM chat
    			 WHERE id = $1 AND userid = $2`,
				messageid, userId,
			)

		} else {
			fmt.Println("403 Forbidden")
			response.Setversion(request.version)
			response.Setcode("403")
			response.Setmessage("Forbidden")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

	} else {

		GivenXSRF := ParseXSRF(request.Headers)

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrf_token
			FROM guest_sessions
			WHERE userid = $1 `,
			hashedGuestString,
		)

		var storedXSRF string
		err = row.Scan(&storedXSRF)

		if err != nil || GivenXSRF != storedXSRF {
			response.Setversion(request.version)
			response.Setmessage("Forbidden")
			response.Setcode("403")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

		row = router.DB.QueryRow(
			context.Background(),
			`SELECT author, id, content, updated, userid
			FROM chat
			WHERE id = $1`,
			messageid,
		)
		err := row.Scan(&author, &id, &content, &updated, &userId)
		if err != nil {
			fmt.Println("Error deleting message", err)
			return
		}

		if hashedGuestString != userId {
			fmt.Println("403 Forbidden")
			response.Setversion(request.version)
			response.Setcode("403")
			response.Setmessage("Forbidden")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

		_, err = router.DB.Exec(
			context.Background(),
			`DELETE FROM chat
		 	WHERE id = $1 AND userid = $2`,
			messageid, hashedGuestString,
		)

		if err != nil {
			fmt.Println("Error deleting message", err)
			return
		}
	}
	response.Setversion(request.version)
	response.Setcode("200")
	response.Setmessage("OK")
	response.Setnosniffheader()
	response.Setbodytext("")
	response.Setcontentlengthheader(0)
	response.Setallbytes()
}

func (router *Router) UpdateMessage(request Request, response *Response) {
	parts := strings.Split(request.path, "/")
	messageid := parts[3]
	authToken := ParseHeaders(request.Headers)
	hashedGuest := sha256.Sum256([]byte(authToken))
	hashedGuestString := hex.EncodeToString(hashedGuest[:])

	var author string
	var id string
	var content string
	var updated bool
	var userId string
	var username string
	var userid string
	var authtoken string

	hashedToken := sha256.Sum256([]byte(authToken))
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT username, userid, authtoken
		 FROM users
		 WHERE authtoken = $1`,
		hashedTokenString,
	)
	err := row.Scan(&username, &userid, &authtoken)
	if err == nil {

		GivenXSRF := ParseXSRF(request.Headers)

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrftoken
			FROM users
			WHERE userid = $1`,
			userid,
		)

		var storedXSRF string
		err = row.Scan(&storedXSRF)

		if err != nil || GivenXSRF != storedXSRF {
			response.Setversion(request.version)
			response.Setmessage("Forbidden")
			response.Setcode("403")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

		row = router.DB.QueryRow(
			context.Background(),
			`SELECT author, id, content, updated, userid FROM chat WHERE id = $1`,
			messageid,
		)
		err = row.Scan(&author, &id, &content, &updated, &userId)
		updated = true
		if err == nil {

			content = html.EscapeString(ParseBody(string(request.body)))
			_, err = router.DB.Exec(
				context.Background(),
				`UPDATE chat
			 	 SET content = $1, updated = $2
    			 WHERE id = $3 AND userid = $4`,
				content, updated, messageid, userid,
			)

		} else {
			fmt.Println("403 Forbidden")
			response.Setversion(request.version)
			response.Setcode("403")
			response.Setmessage("Forbidden")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

	} else {

		GivenXSRF := ParseXSRF(request.Headers)

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrf_token
			FROM guest_sessions
			WHERE userid = $1`,
			hashedGuestString,
		)

		var storedXSRF string
		err = row.Scan(&storedXSRF)

		if err != nil || GivenXSRF != storedXSRF {
			response.Setversion(request.version)
			response.Setmessage("Forbidden")
			response.Setcode("403")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

		row = router.DB.QueryRow(
			context.Background(),
			`SELECT author, id, content, updated, userid FROM chat WHERE id = $1`,
			messageid,
		)

		err = row.Scan(&author, &id, &content, &updated, &userId)
		if err != nil {
			fmt.Println("Error finding message", err)
			return
		}

		if hashedGuestString != userId {
			fmt.Println("403 Forbidden")
			response.Setversion(request.version)
			response.Setcode("403")
			response.Setmessage("Forbidden")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		content = html.EscapeString(ParseBody(string(request.body)))

		_, err = router.DB.Exec(
			context.Background(),
			`UPDATE chat
		 	SET content = $1, updated = true
		 	WHERE id = $2`,
			content,
			messageid,
		)

		if err != nil {
			fmt.Println("Error updating message", err)
			return
		}
	}

	response.Setversion(request.version)
	response.Setcode("200")
	response.Setmessage("OK")
	response.Setnosniffheader()
	response.Setbodytext("")
	response.Setcontentlengthheader(0)
	response.Setallbytes()

}

func (router *Router) GetMessages(request Request, response *Response) {

	rows, err := router.DB.Query(
		context.Background(),
		`SELECT chat.author, chat.id, chat.content, chat.updated,
            chat.userid,
            COALESCE(users.propic, '/public/imgs/user.webp')
     FROM chat
     LEFT JOIN users ON chat.userid = users.userid`,
	)
	if err != nil {
		fmt.Println(err, "Error querying Messages")
		return
	} else {
		defer rows.Close()
		allMessages := []ChatMessage{}
		for rows.Next() {
			var message ChatMessage
			rows.Scan(&message.Author, &message.Id, &message.Content, &message.Updated, &message.Userid, &message.Propic)

			allMessages = append(allMessages, message)
		}
		//make sure i can use marshal
		data, err := json.Marshal(allMessages)
		fullData := `{"messages":` + string(data) + `}`
		if err != nil {
			fmt.Println("Error turning data into json")
			return
		}
		//build response
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		response.Setnosniffheader()
		response.Setbodyjson(string(fullData))
		response.Setcontentlengthheader(len(fullData))
		response.Setallbytes()

	}

}

func (router *Router) ChatHandler(request Request, response *Response) {
	var err error
	var author string
	var id string
	var content string
	var updated bool
	var userId string
	var username string
	var userid string
	var authtoken string

	existingId := ParseHeaders(request.Headers)
	hashedToken := sha256.Sum256([]byte(existingId))
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT username, userid, authtoken
		 FROM users
		 WHERE authtoken = $1`,
		hashedTokenString,
	)
	err = row.Scan(&username, &userid, &authtoken)
	updated = false
	id = string(gofakeit.ID())
	if err == nil {
		GivenXSRF := ParseXSRF(request.Headers)

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrftoken
    		 FROM users
    		 WHERE authtoken = $1`,
			hashedTokenString,
		)

		var storedXSRF string
		err = row.Scan(&storedXSRF)
		if err != nil {
			fmt.Println("error looking for xsrf token", err)
			return
		}

		if err != nil || GivenXSRF != storedXSRF {
			response.Setversion(request.version)
			response.Setmessage("Forbidden")
			response.Setcode("403")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}
		content = html.EscapeString(ParseBody(string(request.body)))
		_, err = router.DB.Exec(
			context.Background(),
			`INSERT INTO chat (author, id, content, updated, userId, claimed)
    		 VALUES ($1, $2, $3, $4, $5, true)`, username, id, content, updated, userid,
		)

		text := "Message Sent"
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		response.Setheaders([]map[string]string{{"Content-Type": "text/html"}})
		response.Setnosniffheader()
		response.Setbodytext(text)
		response.Setcontentlengthheader(len(text))
		response.Setallbytes()
		return

	} else if existingId != "" {
		err = nil
		hashedGuest := sha256.Sum256([]byte(existingId))
		hashedGuestString := hex.EncodeToString(hashedGuest[:])

		GivenXSRF := ParseXSRF(request.Headers)

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrf_token, author
			FROM guest_sessions
			WHERE userid = $1 `,
			hashedGuestString,
		)

		var storedXSRF string
		err = row.Scan(&storedXSRF, &author)
		if err != nil {
			fmt.Println("error looking for xsrf token", err)
			return

		}
		if err != nil || GivenXSRF != storedXSRF {
			response.Setversion(request.version)
			response.Setmessage("Forbidden")
			response.Setcode("403")
			response.Setnosniffheader()
			response.Setbodytext("")
			response.Setcontentlengthheader(0)
			response.Setallbytes()
			return
		}

		row = router.DB.QueryRow(
			context.Background(),
			`SELECT author, id, content, updated, userId 
			FROM chat WHERE userId = $1 AND claimed = false`,
			hashedGuestString,
		)

		err = row.Scan(&author, &id, &content, &updated, &userId)

		if err == pgx.ErrNoRows {
			id = string(gofakeit.ID())
			content = ParseBody(string(request.body))
			updated = false

			content = html.EscapeString(ParseBody(string(request.body)))
			_, err = router.DB.Exec(
				context.Background(),
				`INSERT INTO chat (author, id, content, updated, userId)
    			 VALUES ($1, $2, $3, $4, $5)`, author, id, content, updated, hashedGuestString,
			)

			if err != nil {
				// send err response
				fmt.Println(err, "1error creating/ adding sql table")
			}

		} else {
			err = nil
			id = string(gofakeit.ID())
			content = html.EscapeString(ParseBody(string(request.body)))

			_, err = router.DB.Exec(
				context.Background(),
				`INSERT INTO chat (author, id, content, updated, userId)
    			VALUES ($1, $2, $3, $4, $5)`,
				author,
				id,
				content,
				updated,
				userId,
			)

			if err != nil {
				// send err response
				fmt.Println(err, "1error creating/ adding sql table")
				return
			}
		}
	} else {

		err = nil
		author = gofakeit.Name()
		id = string(gofakeit.ID())
		content = ParseBody(string(request.body))
		updated = false
		randomBytes := make([]byte, 16)
		_, err = rand.Read(randomBytes)
		if err != nil {
			return
		}
		userId = hex.EncodeToString(randomBytes)

		hashedGuest := sha256.Sum256([]byte(userId))
		hashedGuestString := hex.EncodeToString(hashedGuest[:])

		content = html.EscapeString(ParseBody(string(request.body)))
		_, err = router.DB.Exec(
			context.Background(),
			`INSERT INTO chat (author, id, content, updated, userId)
    		VALUES ($1, $2, $3, $4, $5)`, author, id, content, updated, hashedGuestString,
		)
		if err != nil {
			return
		}
	}
	if err != nil {
		// send err response
		fmt.Println(err, "error creating/ adding sql table")
		return
	} else {

		//response double check if we need all this info
		text := "Message Sent"
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		response.Setheaders([]map[string]string{{"Content-Type": "text/html"}})
		response.Setnosniffheader()
		response.Setbodytext(text)
		response.Setcontentlengthheader(len(text))
		response.Setallbytes()
	}

}

func (router *Router) GetFile(request Request, response *Response) {
	existingCookie := ParseHeaders(request.Headers)
	var XSRFTokenstring string
	cookieValue := ""
	hashedToken := sha256.Sum256([]byte(existingCookie))
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	XSRF := make([]byte, 16)
	path := request.path
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT xsrftoken FROM users WHERE authtoken = $1`,
		hashedTokenString,
	)
	err := row.Scan(&XSRFTokenstring)

	if err == nil {
		//user authenticated
		cookieValue = existingCookie
	} else if existingCookie == "" {
		//no cookie, create cookie and XSRF

		cookie := make([]byte, 16)

		_, err := rand.Read(cookie)

		if err != nil {
			fmt.Println("error generating  token", err)
			return
		}
		_, err = rand.Read(XSRF)

		if err != nil {
			fmt.Println("error generating  token", err)
			return
		}
		cookieValue = hex.EncodeToString(cookie)
		hashedCookie := sha256.Sum256([]byte(cookieValue))
		// hashedGuestXSRF := sha256.Sum256([]byte(XSRF))
		newcookie := hex.EncodeToString(hashedCookie[:])
		XSRFToken := hex.EncodeToString(XSRF)
		author := gofakeit.Name()
		_, err = router.DB.Exec(
			context.Background(),
			`INSERT INTO guest_sessions (userid, xsrf_token, author)
    		 VALUES ($1, $2, $3)`,
			newcookie, XSRFToken, author,
		)

		if err != nil {
			// send err response
			fmt.Println(err, "adding info to sql table")
			return
		}
	} else {
		// is a cookie
		var storedtoken string

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrf_token
         	FROM guest_sessions
         	WHERE userid = $1`,
			hashedTokenString,
		)

		err := row.Scan(&storedtoken)
		if err == nil {
			XSRF, err = hex.DecodeString(storedtoken)
			if err != nil {
				return
			}
		}

		if err == pgx.ErrNoRows {
			cookie := make([]byte, 16)
			_, err := rand.Read(cookie)

			if err != nil {
				fmt.Println("error generating  token", err)
				return
			}
			_, err = rand.Read(XSRF)

			if err != nil {
				fmt.Println("error generating  token", err)
				return
			}
			cookieValue = hex.EncodeToString(cookie)
			hashedCookie := sha256.Sum256([]byte(cookieValue))
			// hashedGuestXSRF := sha256.Sum256([]byte(XSRF))
			newcookie := hex.EncodeToString(hashedCookie[:])
			XSRFToken := hex.EncodeToString(XSRF)

			author := gofakeit.Name()

			_, err = router.DB.Exec(
				context.Background(),
				`INSERT INTO guest_sessions (userid, xsrf_token, author)
     VALUES ($1, $2, $3)`,
				newcookie, XSRFToken, author,
			)

			if err != nil {
				// send err response
				fmt.Println(err, "adding info to sql table")
				return
			}

		}
		if err == nil {
			cookieValue = existingCookie
		}
	}

	if path == "/" {
		path = "public/index.html"
	} else {
		path = "public" + path + ".html"
	}
	if path == "public/videotube/upload.html" {
		path = "public/upload.html"
	} else if strings.HasPrefix(path, "public/videotube/videos/") {
		path = "public/view-video.html"
	}
	body, err := os.ReadFile(path)
	browserEtag := ParseEtag(request.Headers)
	hash := sha256.Sum256(body)
	hashString := fmt.Sprintf("%x", hash)
	if err != nil {
		//go to 404 page
		body, _ := os.ReadFile("public/404.html")
		//build response
		response.Setversion(request.version)
		response.Setmessage("Not Found")
		response.Setheaders([]map[string]string{
			{"Content-Type": "text/html"},
		})
		layout, _ := os.ReadFile("public/layout/layout.html")
		rendered := strings.Replace(
			string(layout),
			"{{content}}",
			string(body),
			1,
		)
		response.Setversion(request.version)
		response.Setmessage("Not Found")
		response.Setcode("404")
		response.Setnosniffheader()
		response.Setbodytext(rendered)
		response.Setcontentlengthheader(len(rendered))
		response.Setallbytes()
		return
	}
	// no error build response

	response.Setversion(request.version)
	response.Setmessage("OK")
	response.Setcode("200")
	extension := GetMimeType(path)
	// changed this so dobule check
	if browserEtag != hashString || extension == "html" {
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		if extension == "jpg" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/jpeg"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "png" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/png"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodytext(string(body))
			response.Setcontentlengthheader(len(body))
		} else if extension == "js" {
			response.Setheaders([]map[string]string{{"Content-Type": "text/javascript"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodytext(string(body))
			response.Setcontentlengthheader(len(body))
		} else if extension == "ico" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/x-icon"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "gif" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/gif"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "webp" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/webp"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "mp4" {
			response.Setheaders([]map[string]string{{"Content-Type": "video/mp4"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}, {"Accept-Ranges": "bytes"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "html" {
			layout, _ := os.ReadFile("public/layout/layout.html")
			rendered := strings.Replace(
				string(layout),
				"{{content}}",
				string(body),
				1,
			)
			if XSRFTokenstring == "" {
				XSRFTokenstring = hex.EncodeToString(XSRF)
			}
			htmlContent := strings.ReplaceAll(string(rendered), "{{XSRF_TOKEN}}", XSRFTokenstring)
			response.Setheaders([]map[string]string{{"Content-Type": "text/html"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodytext(htmlContent)
			response.Setcontentlengthheader(len(htmlContent))
		}
		response.Setallbytes()
	} else {
		response.Setversion(request.version)
		response.Setmessage("Not Modified")
		response.Setcode("304")
		response.Setnosniffheader()
		response.Setallbytes()
	}

	// if extension == "jpg" {
	// 	//later
	// 	return
	// } else if extension == "ico" {

	// 	//later
	// 	return
	// } else if extension == "mp4" {
	// 	response.Setheaders([]map[string]string{{"Content-Type": "video/mp4"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
	// 	response.Setnosniffheader()
	// 	response.Setbodybinary(body)
	// 	response.Setcontentlengthheader(len(body))

	// } else if extension == "gif" {
	// 	//later
	// 	return
	// } else if extension == "webp" {
	// 	//later
	// 	return
	// 	// add png to if adding these others
	// } else if extension == "html" {
	// 	layout, _ := os.ReadFile("public/layout/layout.html")

	// 	rendered := strings.Replace(
	// 		string(layout),
	// 		"{{content}}",
	// 		string(body),
	// 		1,
	// 	)
	// 	if XSRFTokenstring == "" {
	// 		XSRFTokenstring = hex.EncodeToString(XSRF)
	// 	}
	// 	htmlContent := strings.ReplaceAll(string(rendered), "{{XSRF_TOKEN}}", XSRFTokenstring)
	// 	response.Setheaders([]map[string]string{{"Content-Type": "text/html"}, {"Set-Cookie": cookieValue + ";Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
	// 	response.Setnosniffheader()
	// 	response.Setbodytext(htmlContent)
	// 	response.Setcontentlengthheader(len(htmlContent))

	// }
	// response.Setallbytes()
}

func RangeRequest(rangeValue string, requestpath string, request Request, response *Response) {
	var start int
	var finish int
	finishStr := ""
	parts := strings.Split(rangeValue, "=")
	if len(parts) != 2 || !strings.Contains(parts[1], "-") {
		response.Setversion(request.version)
		response.Setmessage("Range Not Satisfiable")
		response.Setcode("416")
		response.Setnosniffheader()
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}

	num := strings.Split(parts[1], "-")
	if num[0] != "" && num[1] != "" {
		start, _ = strconv.Atoi(num[0])
		finish, _ = strconv.Atoi(num[1])
	} else if num[0] != "" && num[1] == "" {
		start, _ = strconv.Atoi(num[0])
		finishStr = "tilend"
	} else if num[0] == "" && num[1] != "" {
		start, _ = strconv.Atoi(num[1])
		finishStr = "lastbytes"
	}

	file, err := os.Open(requestpath)
	if err != nil {
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return
	}
	size := int(info.Size())

	if finishStr == "lastbytes" {
		start = size - start
		if start < 0 {
			start = 0
		}
		finish = size - 1
	} else if finishStr == "tilend" {
		finish = size - 1
	}
	if start >= size || start > finish {
		response.Setversion(request.version)
		response.Setmessage("Range Not Satisfiable")
		response.Setcode("416")
		response.Setnosniffheader()
		response.Setheaders([]map[string]string{{"Content-Range": "bytes */" + strconv.Itoa(size)}})
		response.Setbodytext("")
		response.Setcontentlengthheader(0)
		response.Setallbytes()
		return
	}
	if finish >= size {
		finish = size - 1
	}

	file.Seek(int64(start), io.SeekStart)
	videoBytes := make([]byte, finish-start+1)
	_, err = io.ReadFull(file, videoBytes)
	if err != nil {
		return
	}
	response.Setversion(request.version)
	response.Setmessage("Partial Content")
	response.Setcode("206")
	response.Setheaders([]map[string]string{{"Content-Type": "video/mp4"}, {"Accept-Ranges": "bytes"}, {"Content-Range": fmt.Sprintf("bytes %d-%d/%d", start, finish, size)}})
	response.Setnosniffheader()
	response.Setbodybinary(videoBytes)
	response.Setcontentlengthheader(len(videoBytes))
	response.Setallbytes()
}

func (router *Router) GetPublicFile(request Request, response *Response) {

	existingCookie := ParseHeaders(request.Headers)
	var XSRFTokenstring string

	cookieValue := ""
	hashedToken := sha256.Sum256([]byte(existingCookie))
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	XSRF := make([]byte, 16)

	row := router.DB.QueryRow(
		context.Background(),
		`SELECT xsrftoken FROM users WHERE authtoken = $1`,
		hashedTokenString,
	)
	err := row.Scan(&XSRFTokenstring)

	if err == nil {
		//user authenticated
		cookieValue = existingCookie
	} else if existingCookie == "" {
		//no cookie, create cookie and XSRF

		cookie := make([]byte, 16)

		_, err := rand.Read(cookie)

		if err != nil {
			fmt.Println("error generating  token", err)
			return
		}
		_, err = rand.Read(XSRF)

		if err != nil {
			fmt.Println("error generating  token", err)
			return
		}
		cookieValue = hex.EncodeToString(cookie)
		hashedCookie := sha256.Sum256([]byte(cookieValue))
		newcookie := hex.EncodeToString(hashedCookie[:])
		XSRFToken := hex.EncodeToString(XSRF)
		author := gofakeit.Name()
		_, err = router.DB.Exec(
			context.Background(),
			`INSERT INTO guest_sessions (userid, xsrf_token, author)
    		 VALUES ($1, $2, $3)`,
			newcookie, XSRFToken, author,
		)

		if err != nil {
			// send err response
			fmt.Println(err, "adding info to sql table")
			return
		}
	} else {
		// is a cookie
		var storedtoken string

		row := router.DB.QueryRow(
			context.Background(),
			`SELECT xsrf_token
         	FROM guest_sessions
         	WHERE userid = $1`,
			hashedTokenString,
		)

		err := row.Scan(&storedtoken)
		if err == nil {
			XSRF, err = hex.DecodeString(storedtoken)
			if err != nil {
				return
			}
		}

		if err == pgx.ErrNoRows {
			cookie := make([]byte, 16)
			_, err := rand.Read(cookie)

			if err != nil {
				fmt.Println("error generating  token", err)
				return
			}
			_, err = rand.Read(XSRF)

			if err != nil {
				fmt.Println("error generating  token", err)
				return
			}
			cookieValue = hex.EncodeToString(cookie)
			hashedCookie := sha256.Sum256([]byte(cookieValue))
			newcookie := hex.EncodeToString(hashedCookie[:])
			XSRFToken := hex.EncodeToString(XSRF)

			author := gofakeit.Name()

			_, err = router.DB.Exec(
				context.Background(),
				`INSERT INTO guest_sessions (userid, xsrf_token, author)
     			VALUES ($1, $2, $3)`,
				newcookie, XSRFToken, author,
			)

			if err != nil {
				// send err response
				fmt.Println(err, "adding info to sql table")
				return
			}

		}
		if err == nil {
			cookieValue = existingCookie
		}
	}

	extension := GetMimeType(request.path)
	// if extension == "html" {
	// 	path = request.path + ".html"
	// } else {
	// 	path = request.path
	// }
	// check that all these public file paths work after change
	if request.path == "public/videotube/upload.html" {
		request.path = "public/upload.html"
	} else if strings.HasPrefix(request.path, "public/videotube/videos/") {
		request.path = "public/view-video.html"
	}
	var body []byte

	rangeHeader := ParseRange(request.Headers)
	if rangeHeader != "" && extension == "mp4" {
		RangeRequest(rangeHeader, strings.TrimPrefix(request.path, "/"), request, response)
		return
	} else {
		body, err = os.ReadFile(strings.TrimPrefix(request.path, "/"))
	}
	browserEtag := ParseEtag(request.Headers)
	hash := sha256.Sum256(body)
	hashString := fmt.Sprintf("%x", hash)

	if err != nil {
		//go to 404 page
		fmt.Println("error find page", err)
		body, _ := os.ReadFile("public/404.html")
		//build response
		response.Setheaders([]map[string]string{
			{"Content-Type": "text/html"},
		})
		layout, _ := os.ReadFile("public/layout/layout.html")
		rendered := strings.Replace(
			string(layout),
			"{{content}}",
			string(body),
			1,
		)
		response.Setversion(request.version)
		response.Setmessage("Not Found")
		response.Setcode("404")
		response.Setnosniffheader()
		response.Setbodytext(rendered)
		response.Setcontentlengthheader(len(rendered))
		response.Setallbytes()
		return
	}
	if browserEtag != hashString || extension == "html" {
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		if extension == "jpg" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/jpeg"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "png" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/png"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodytext(string(body))
			response.Setcontentlengthheader(len(body))
		} else if extension == "js" {
			response.Setheaders([]map[string]string{{"Content-Type": "text/javascript"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodytext(string(body))
			response.Setcontentlengthheader(len(body))
		} else if extension == "ico" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/x-icon"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "gif" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/gif"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "webp" {
			response.Setheaders([]map[string]string{{"Content-Type": "image/webp"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "mp4" {
			response.Setheaders([]map[string]string{{"Content-Type": "video/mp4"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}, {"Accept-Ranges": "bytes"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "m3u8" {
			response.Setheaders([]map[string]string{{"Content-Type": "application/vnd.apple.mpegurl"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodytext(string(body))
			response.Setcontentlengthheader(len(body))
		} else if extension == "ts" {
			response.Setheaders([]map[string]string{{"Content-Type": "video/mp2t"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodybinary(body)
			response.Setcontentlengthheader(len(body))
		} else if extension == "html" {
			layout, _ := os.ReadFile("public/layout/layout.html")
			rendered := strings.Replace(
				string(layout),
				"{{content}}",
				string(body),
				1,
			)
			if XSRFTokenstring == "" {
				XSRFTokenstring = hex.EncodeToString(XSRF)
			}
			htmlContent := strings.ReplaceAll(string(rendered), "{{XSRF_TOKEN}}", XSRFTokenstring)
			response.Setheaders([]map[string]string{{"Content-Type": "text/html"}, {"Cache-Control": "max-age=3600"}, {"ETag": hashString}, {"Set-Cookie": cookieValue + "; Max-Age=3600; HttpOnly; SameSite=Lax; Path=/"}})
			response.Setnosniffheader()
			response.Setbodytext(htmlContent)
			response.Setcontentlengthheader(len(htmlContent))
		}
		response.Setallbytes()
	} else {
		response.Setversion(request.version)
		response.Setmessage("Not Modified")
		response.Setcode("304")
		response.Setnosniffheader()
		response.Setallbytes()
	}
}

func GetMimeType(path string) string {
	parts := strings.Split(path, ".")
	extension := parts[len(parts)-1]
	return extension
}
