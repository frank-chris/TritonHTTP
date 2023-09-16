package tritonhttp

import (
	"io"
	"os"
	"sort"
	"strconv"
)

type Response struct {
	Proto      string // e.g. "HTTP/1.1"
	StatusCode int    // e.g. 200
	StatusText string // e.g. "OK"

	// Headers stores all headers to write to the response.
	Headers map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	// Hint: you might need this to handle the "Connection: Close" requirement
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

func (response *Response) writeResponse(writer io.Writer) error {
	delim := "\r\n"
	// Status
	statusCode := strconv.Itoa(response.StatusCode)
	statusLine := response.Proto + " " + statusCode + " " + response.StatusText + delim

	_, err := writer.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	// Headers
	keys := make([]string, 0, len(response.Headers))
	for key := range response.Headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		header := key + ": " + response.Headers[key] + delim
		_, err := writer.Write([]byte(header))
		if err != nil {
			return err
		}
	}
	_, err = writer.Write([]byte(delim))
	if err != nil {
		return err
	}

	// Body
	body := []byte{}
	if response.FilePath != "" {
		body, err = os.ReadFile(response.FilePath)
		if err != nil {
			return err
		}
	}
	_, err = writer.Write(body)
	if err != nil {
		return err
	}
	return nil
}
