package utils

import (
	_"fmt"
	"strings"
)

type Request struct {
	method  string
	path    string
	headers []map[string]string
	body    string
	version string
}

// will have to change this for hw 3 to update parsing images/videos
func Parse(request string, requestObj *Request) {
	requestObj.headers = nil
	i := 1
	isBody := false
	parts := strings.Split(request, "\r\n")
	initialsplit := strings.Split(parts[0], " ")
	requestObj.method = initialsplit[0]
	requestObj.path = initialsplit[1]
	requestObj.version = initialsplit[2]
	length := len(parts)
	// not the best with nested loops but headers will never be that big
	for ; i < length - 1; i++ {
		key, value, _ := strings.Cut(parts[i], ":")
		space := value[0]
		if space == ' ' {
			value = value [1:]

		}
		requestObj.headers = append(requestObj.headers, map[string]string{key: value})
		
		if isBody != true {
			for j := range requestObj.headers {
				for key, _ := range requestObj.headers[j] {
					if strings.ToLower(key) == "content-length" || strings.ToLower(key) == "transfer-encoding" {
						isBody = true
						length--
						break
					}
				}
			}
		}
		// got confused here with i 
		if i + 3 == length && isBody != true{
			break;
		}
		
	}

	if isBody {
		requestObj.body = parts[i + 1]
	}

	
}
