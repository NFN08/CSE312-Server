package utils

import (

	_"fmt"
	"strings"
	"bytes"
	"net"
)

type Request struct {
	method  string
	path    string
	Headers []map[string]string
	body    []byte
	version string
	Conn net.Conn
}

// will have to change this for hw 3 to update parsing images/videos
func Parse(request []byte, requestObj *Request) {
	requestObj.Headers = nil
	i := 1
	isBody := false
	// headerSection, body, _ := strings.Cut(request, "\r\n\r\n")
	headerSection, body, _ := bytes.Cut(request, []byte("\r\n\r\n"))
	parts := strings.Split(string(headerSection), "\r\n")
	initialsplit := strings.Split(parts[0], " ")
	requestObj.method = initialsplit[0]
	requestObj.path = initialsplit[1]
	requestObj.version = initialsplit[2]
	length := len(parts)
	// not the best with nested loops but headers will never be that big, maybe fix
	for ; i < length; i++ {
		key, value, _ := strings.Cut(parts[i], ":")
		space := value[0]
		if space == ' ' {
			value = value [1:]

		}
		requestObj.Headers = append(requestObj.Headers, map[string]string{key: value})
		
		if isBody != true {
			for j := range requestObj.Headers {
				for key, _ := range requestObj.Headers[j] {
					if strings.ToLower(key) == "content-length" || strings.ToLower(key) == "transfer-encoding" {
						isBody = true
						break
					}
				
				}
			}
		}		
	}
	if isBody {
		requestObj.body = body
	}
	
	
}
