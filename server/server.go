package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/s4y/reserve"
)

var mu sync.Mutex

func persist(obj interface{}) error {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.OpenFile("data.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

func main() {
	httpAddr := flag.String("http", "127.0.0.1:8025", "Listening address")
	flag.Parse()
	fmt.Printf("http://%s/\n", *httpAddr)

	ln, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", reserve.FileServer("../static"))
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		email := r.FormValue("email")
		fmt.Println(email)
	})
	log.Fatal(http.Serve(ln, nil))
}
