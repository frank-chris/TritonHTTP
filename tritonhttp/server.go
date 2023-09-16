package tritonhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// VirtualHosts contains a mapping from host name to the docRoot path
	// (i.e. the path to the directory to serve static files from) for
	// all virtual hosts that this server supports
	VirtualHosts map[string]string
}

func (server *Server) goodRequest(request *Request) *Response {
	response := &Response{}
	response.Proto = "HTTP/1.1"
	response.Request = request
	response.Headers = make(map[string]string)
	response.Headers["Date"] = FormatTime(time.Now())
	if request.URL[len(request.URL)-1] == '/' {
		request.URL += "index.html"
	}
	if request.Close {
		response.Headers["Connection"] = "close"
	}
	docRoot, validHost := server.VirtualHosts[request.Host]
	if !validHost {
		response.StatusCode = 404
		response.StatusText = "Not Found"
		return response
	}

	path := filepath.Join(docRoot, filepath.Clean(request.URL))

	fileInfo, err := os.Stat(path)

	if !errors.Is(err, os.ErrNotExist) && fileInfo.IsDir() {
		path += "/index.html"
		fileInfo, err = os.Stat(path)
	}

	if errors.Is(err, os.ErrNotExist) || path[:len(docRoot)] != docRoot {
		response.StatusCode = 404
		response.StatusText = "Not Found"
	} else if fileInfo.IsDir() {
		response.StatusCode = 404
		response.StatusText = "Not Found"
	} else {
		response.StatusCode = 200
		response.StatusText = "OK"
		response.FilePath = path
		response.Headers["Last-Modified"] = FormatTime(fileInfo.ModTime())
		response.Headers["Content-Type"] = MIMETypeByExtension(filepath.Ext(path))
		response.Headers["Content-Length"] = strconv.FormatInt(fileInfo.Size(), 10)
	}
	return response
}

func (server *Server) badRequest() *Response {
	response := &Response{}
	response.Proto = "HTTP/1.1"
	response.Request = nil
	response.Headers = make(map[string]string)
	response.Headers["Date"] = FormatTime(time.Now())
	response.StatusCode = 400
	response.StatusText = "Bad Request"
	response.Headers["Connection"] = "close"
	return response
}

func (server *Server) HandleRequests(connection net.Conn) {
	reader := bufio.NewReader(connection)

	for {
		err := connection.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Printf("Error while setting timeout for %v", connection)
			connection.Close()
			return
		}

		request, bytesReceived, err := readRequest(reader)

		// EOF
		if err == io.EOF {
			log.Printf("Connection closed by client %v", connection.RemoteAddr())
			connection.Close()
			return
		}

		// Timeout
		if err, ok := err.(net.Error); ok && err.Timeout() {
			if !bytesReceived {
				log.Printf("Connection to %v timed out", connection.RemoteAddr())
				connection.Close()
				return
			}
			response := server.badRequest()
			err := response.writeResponse(connection)
			if err != nil {
				log.Println(err)
			}
			connection.Close()
			return
		}

		// Bad request
		if err != nil {
			log.Printf("Bad request, error: %v", err)
			response := server.badRequest()
			err := response.writeResponse(connection)
			if err != nil {
				log.Println(err)
			}
			connection.Close()
			return
		}

		// Good request
		log.Printf("Good request: %v", request)
		response := server.goodRequest(request)
		err = response.writeResponse(connection)
		if err != nil {
			log.Println(err)
		}

		// Close
		if request.Close {
			connection.Close()
			return
		}
	}
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (server *Server) ListenAndServe() error {

	// Hint: Validate all docRoots
	for _, docRoot := range server.VirtualHosts {
		fileInfo, err := os.Stat(docRoot)
		if os.IsNotExist(err) {
			return err
		}
		if !fileInfo.IsDir() {
			return fmt.Errorf("doc root %q not a directory", docRoot)
		}
	}

	// Hint: create your listen socket and spawn off goroutines per incoming client
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}

	log.Printf("Server listening on %q", listener.Addr())

	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			return err
		}
		log.Printf("Accepted %q", connection.RemoteAddr())
		go server.HandleRequests(connection)
	}
}
