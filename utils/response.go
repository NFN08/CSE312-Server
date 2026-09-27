package utils
import "fmt"
type Response struct{
	version string
	code string
	message string
	headers []map[string]string
	bodytext string
	bodyjson string
	bodybinary []byte
	Bytes []byte
	contentlengthheader int

}

func (response *Response) Setversion(version string){
	response.version = version
}

// maybe set default code to 200 OK
func (response *Response) Setcode(code string){
	response.code = code
}

func (response *Response) Setmessage(message string){
	response.message = message
}

func (response *Response) Setheaders(headers []map[string]string){
	response.headers = headers

}

func (response *Response) Setbodytext(body string){
	response.bodytext = body
}
func (response *Response) Setbodyjson(body string){
	response.bodyjson = body
	 response.headers = append(response.headers, map[string]string{
        "Content-Type": "application/json",
    })
}
func (response *Response) Setbodybinary(body []byte){
	response.bodybinary = body
}
func (response *Response) Setallbytes (){
	CRLF := "\r\n"
	bytes := []byte(response.version + " " + string(response.code) +" " + response.message + CRLF)
	for _, header := range response.headers{
		for key, value := range header{
			bytes = append(bytes, []byte(key + ":" + value + CRLF)...)
		}
	}
	bytes = append(bytes, CRLF...)
	if response.bodytext != ""{
		bytes = append(bytes, []byte(response.bodytext)...)
	}else if response.bodyjson != ""{
		bytes = append(bytes, []byte(response.bodyjson)...)
	}else if len(response.bodybinary) > 0{
		bytes = append(bytes, []byte(response.bodybinary)...)
	}
	response.Bytes = bytes
}

func (response *Response) Setcontentlengthheader (length int){
	response.contentlengthheader = length
	 response.headers = append(response.headers, map[string]string{
        "Content-Length": fmt.Sprintf("%d", length),
    })
}
// make sure this is always called
func (response *Response) Setnosniffheader() {
    response.headers = append(response.headers, map[string]string{
        "X-Content-Type-Options": "nosniff",
    })
}

