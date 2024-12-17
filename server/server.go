package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/s4y/reserve"
)

type connList struct {
	connections []*websocket.Conn
	lock        sync.RWMutex
}

var conns connList

func broadcast(msg interface{}) {
	conns.lock.RLock()
	for _, conn := range conns.connections {
		conn.WriteJSON(msg)
	}
	defer conns.lock.RUnlock()
}

type persistFile struct {
	mu       sync.RWMutex
	filename string
}

var rsvpFile = &persistFile{filename: "../data.json"}
var videosFile = &persistFile{filename: "../videoSubmissions.json"}

func persist(f *persistFile, obj interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return nil
}

func dump(f *persistFile) ([]interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, err := os.OpenFile(f.filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	ret := []interface{}{}
	for scanner.Scan() {
		var val interface{}
		err := json.Unmarshal(scanner.Bytes(), &val)
		if err != nil {
			return nil, err
		}
		ret = append(ret, val)
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return ret, nil
}

func main() {
	httpAddr := flag.String("http", "127.0.0.1:8025", "Listening address")
	flag.Parse()
	fmt.Printf("http://%s/\n", *httpAddr)

	ln, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		log.Fatal(err)
	}

	upgrader := websocket.Upgrader{}

	http.Handle("/", reserve.FileServer("../static"))
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		persist(rsvpFile, struct {
			Timestamp string `json:"timestamp"`
			IP        string `json:"ip"`
			Email     string `json:"email"`
		}{time.Now().UTC().String(), r.RemoteAddr, r.FormValue("email")})
	})

	http.HandleFunc("/submitVideo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		text := r.FormValue("text")

		if len(text) > 1024 {
			http.Error(w, "that's a little too big", http.StatusRequestEntityTooLarge)
			return
		}

		msg := struct {
			Timestamp string `json:"timestamp"`
			Text      string `json:"text"`
		}{time.Now().UTC().String(), text}

		persist(videosFile, msg)
		broadcast(msg)
	})

	handleErr := func(err error, w http.ResponseWriter) {
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}

	http.HandleFunc("/videoSubmissions", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			handleErr(err, w)
			return
		}

		conns.lock.Lock()
		conns.connections = append(conns.connections, conn)
		conns.lock.Unlock()

		defer func() {
			conns.lock.Lock()
			defer conns.lock.Unlock()
			for i, cur_conn := range conns.connections {
				if cur_conn == conn {
					conns.connections = append(conns.connections[:i], conns.connections[i+1:]...)
					break
				}
			}
		}()

		lines, err := dump(videosFile)
		if err != nil {
			handleErr(err, w)
			return
		}

		for _, line := range lines {
			conn.WriteJSON(line)
		}

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}

	})
	log.Fatal(http.Serve(ln, nil))
}
