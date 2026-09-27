package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	"html"
	"os"
	"strings"
)

type Route struct {
	method string
	path   string
	toCall func(Request, *Response)
}
type Router struct {
	route []Route
	DB    *pgx.Conn
}
type ChatMessage struct {
	Author  string `json:"author"`
	Id      string `json:"id"`
	Content string `json:"content"`
	Updated bool   `json:"updated"`
	Userid  string `json:userid`
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
	if len(request.path) >= len("/api/chats/") {
		chatPath = request.path[:len("/api/chats/")]
	}
	if publicPath != "/public" {
		parts := strings.Split(request.path, ".")
		path := parts[0]
		request.path = path
	}
	for _, route := range router.route {

		if route.method == request.method && route.path == request.path {
			route.toCall(*request, response)
			return

		} else if route.method == request.method && publicPath == route.path {
			route.toCall(*request, response)
			return

		} else if route.method == request.method && chatPath == route.path {
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
	id := ""
	for j := range request {
		for key, value := range request[j] {
			if key == "Cookie" {
				parts := strings.Split(value, "=")
				id = parts[1]
				return id
			}
		}
	}
	return ""
}

func ParseConnection(request []map[string]string) bool {
	close := false
	for j := range request {
		for key, value := range request[j] {
			if key == "Connection" && value == "close" {
				close = true
				return close
			}
		}
	}
	return false
}

func (router *Router) DeleteMessage(request Request, response *Response) {
	parts := strings.Split(request.path, "/")
	messageid := parts[3]
	authToken := ParseHeaders(request.Headers)

	var author string
	var id string
	var content string
	var updated bool
	var userId string
	row := router.DB.QueryRow(
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

	if authToken != userId {
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
		 WHERE id = $1`,
		messageid,
	)

	if err != nil {
		fmt.Println("Error deleting message", err)
		return
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
	var author string
	var id string
	var content string
	var updated bool
	var userId string
	row := router.DB.QueryRow(
		context.Background(),
		`SELECT * FROM chat WHERE id = $1`,
		messageid,
	)

	err := row.Scan(&author, &id, &content, &updated, &userId)
	if err != nil {
		fmt.Println("Error finding message", err)
		return
	}

	if authToken != userId {
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
	content = html.EscapeString(ParseBody(request.body))

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
		`SELECT * FROM chat`,
	)
	if err != nil {
		fmt.Println(err, "Error querying Messages")
	} else {

		allMessages := []ChatMessage{}
		for rows.Next() {
			var message ChatMessage
			rows.Scan(&message.Author, &message.Id, &message.Content, &message.Updated, &message.Userid)

			allMessages = append(allMessages, message)
		}
		//make sure i can use marshal
		data, err := json.Marshal(allMessages)
		fullData := `{"messages":` + string(data) + `}`
		if err != nil {
			fmt.Println("Error turning data into json")
		}
		//build response
		fmt.Println("GET RETURNING:", fullData)
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		response.Setnosniffheader()
		response.Setbodytext(string(fullData))
		response.Setcontentlengthheader(len(fullData))
		response.Setallbytes()

	}

}

func (router *Router) ChatHandler(request Request, response *Response) {
	//store message and message id plus author in postgres
	//use prepared statements
	// add timeouts and fallbacks for postgres
	// _, err := router.DB.Exec(
	// 	context.Background(),
	// 	`
	// CREATE TABLE IF NOT EXISTS chat (
	//     author TEXT,
	// 	id TEXT,
	//     content TEXT,
	//     updated BOOLEAN,
	// 	userId TEXT
	// )`)
	var err error
	var author string
	var id string
	var content string
	var updated bool
	var userId string

	// call func to iterate over headers and get set cookie header id and then fetch that id from db and get info
	existingId := ParseHeaders(request.Headers)
	if existingId != "" {
		row := router.DB.QueryRow(
			context.Background(),
			`SELECT * FROM chat WHERE userId = $1`,
			existingId,
		)

		err := row.Scan(&author, &id, &content, &updated, &userId)
		//was having problems when swtichign between docker and reg
		if err == pgx.ErrNoRows {
			author = gofakeit.Name()
			id = string(gofakeit.ID())
			userId = string(gofakeit.ID())
			content = ParseBody(request.body)
			updated = false

			content = html.EscapeString(ParseBody(request.body))
			_, err = router.DB.Exec(
				context.Background(),
				`INSERT INTO chat (author, id, content, updated, userId)
    			 VALUES ($1, $2, $3, $4, $5)`, author, id, content, updated, userId,
			)

		} else {
			id = string(gofakeit.ID())
			content = ParseBody(request.body)

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
				// figure out how to quit here and others
			}
		}
	} else {

		// if user does not exist do this otherwise get id from cookie, fetch from db and insert query with that id/ author

		author = gofakeit.Name()
		id = string(gofakeit.ID())
		userId = string(gofakeit.ID())
		content = ParseBody(request.body)
		updated = false

		content = html.EscapeString(ParseBody(request.body))
		_, err = router.DB.Exec(
			context.Background(),
			`INSERT INTO chat (author, id, content, updated, userId)
    		VALUES ($1, $2, $3, $4, $5)`, author, id, content, updated, userId,
		)
	}
	if err != nil {
		// send err response
		fmt.Println(err, "error creating/ adding sql table")
	} else {

		//response double check if we need all this info
		cookie := fmt.Sprintf("Id=%s; Path=/", userId)
		text := "Message Sent"
		response.Setversion(request.version)
		response.Setmessage("OK")
		response.Setcode("200")
		response.Setheaders([]map[string]string{{"Content-Type": "text/html"}})
		response.Setheaders([]map[string]string{{"Set-Cookie": cookie + ";Max-Age=3600; HttpOnly"}})
		response.Setnosniffheader()
		response.Setbodytext(text)
		response.Setcontentlengthheader(len(text))
		response.Setallbytes()
	}

}

func (router *Router) GetFile(request Request, response *Response) {
	path := request.path
	if path == "/" {
		path = "public/index.html"
	} else {
		path = "public" + path + ".html"
	}
	body, err := os.ReadFile(path)
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
	if extension == "jpg" {
		//later
		return
	} else if extension == "ico" {
		//later
		return
	} else if extension == "gif" {
		//later
		return
	} else if extension == "webp" {
		//later
		return
	} else if extension == "html" {
		layout, _ := os.ReadFile("public/layout/layout.html")

		rendered := strings.Replace(
			string(layout),
			"{{content}}",
			string(body),
			1,
		)

		response.Setheaders([]map[string]string{{"Content-Type": "text/html"}})
		response.Setnosniffheader()
		response.Setbodytext(rendered)
		response.Setcontentlengthheader(len(rendered))

	}
	response.Setallbytes()
}

func (router *Router) GetPublicFile(request Request, response *Response) {
	var path string
	// if strings.Contains(request.path, "js") {
	// 	path = request.path + ".js"
	// }else if strings.Contains(request.path, "jpg") {
	// 	path = request.path + ".jpg"
	// }else if strings.Contains(request.path, "ico") {
	// 	path = request.path + ".ico"
	// }else if strings.Contains(request.path, "webp") {
	// 	path = request.path + ".webp"
	// }else if strings.Contains(request.path, "gif") {
	// 	path = request.path + ".gif"
	// }else {
	// 	path = request.path + ".html"
	// }

	extension := GetMimeType(request.path)
	if extension != "html" {
		path = request.path
	} else {
		path = request.path + ".html"
	}

	body, err := os.ReadFile("." + path)
	if err != nil {
		//go to 404 page
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
	// no error build response
	response.Setversion(request.version)
	response.Setmessage("OK")
	response.Setcode("200")
	// extension := GetMimeType(path)
	if extension == "jpg" {
		response.Setheaders([]map[string]string{{"Content-Type": "image/jpeg"}})
		response.Setnosniffheader()
		response.Setbodybinary(body)
		response.Setcontentlengthheader(len(body))
	} else if extension == "js" {
		response.Setheaders([]map[string]string{{"Content-Type": "text/javascript"}})
		response.Setnosniffheader()
		response.Setbodytext(string(body))
		response.Setcontentlengthheader(len(body))
	} else if extension == "ico" {
		response.Setheaders([]map[string]string{{"Content-Type": "image/x-icon"}})
		response.Setnosniffheader()
		response.Setbodybinary(body)
		response.Setcontentlengthheader(len(body))
	} else if extension == "gif" {
		response.Setheaders([]map[string]string{{"Content-Type": "image/gif"}})
		response.Setnosniffheader()
		response.Setbodybinary(body)
		response.Setcontentlengthheader(len(body))
	} else if extension == "webp" {
		response.Setheaders([]map[string]string{{"Content-Type": "image/webp"}})
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
		response.Setheaders([]map[string]string{{"Content-Type": "text/html"}})
		response.Setnosniffheader()
		response.Setbodytext(rendered)
		response.Setcontentlengthheader(len(rendered))
	}

	response.Setallbytes()
}

func GetMimeType(path string) string {
	parts := strings.Split(path, ".")
	extension := parts[len(parts)-1]

	return extension
}
