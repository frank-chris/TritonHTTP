package tritonhttp

import (
	"bufio"
	"fmt"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Headers stores the key-value HTTP headers
	Headers map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

func readLine(reader *bufio.Reader) (string, error) {
	var line string
	for {
		s, err := reader.ReadString('\n')
		line += s
		if err != nil {
			return line, err
		}
		if strings.HasSuffix(line, "\r\n") {
			line = line[:len(line)-2]
			return line, nil
		}
	}
}

func readRequest(reader *bufio.Reader) (*Request, bool, error) {
	request := &Request{}
	request.Headers = make(map[string]string)

	initialLine, err := readLine(reader)
	if err != nil {
		return nil, false, err
	}

	requestFields := strings.SplitN(initialLine, " ", 3)
	if len(requestFields) != 3 {
		return nil, true, fmt.Errorf("400")
	} else {
		request.Method = requestFields[0]
		request.URL = requestFields[1]
		request.Proto = requestFields[2]
	}

	if request.Method != "GET" || request.URL[0] != '/' || request.Proto != "HTTP/1.1" {
		return nil, true, fmt.Errorf("400")
	}

	request.Close = false
	hostPresent := false
	for {
		line, err := readLine(reader)
		if err != nil {
			return nil, true, err
		}
		if line == "" {
			break
		}

		keyValue := strings.SplitN(line, ": ", 2)
		if len(keyValue) != 2 {
			return nil, true, fmt.Errorf("400")
		}

		key := CanonicalHeaderKey(strings.TrimSpace(keyValue[0]))
		value := strings.TrimSpace(keyValue[1])

		if key == "Host" {
			request.Host = value
			hostPresent = true
		} else if key == "Connection" && value == "close" {
			request.Close = true
		} else {
			request.Headers[key] = value
		}
	}

	if !hostPresent {
		return nil, true, fmt.Errorf("400")
	} else {
		return request, true, nil
	}
}
