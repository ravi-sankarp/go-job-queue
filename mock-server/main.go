package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		n := rand.Intn(100)
		switch {
		case n >= 40:

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))

		case n > 20:

			http.Error(w, "Bad request", http.StatusBadRequest)

		case n == 1:
			select {}
		default:
			delay := time.Duration(1+n) * time.Second
			time.Sleep(delay)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Slow(ok)"))
		}
	})

	log.Println("mock server listening on :8080")
	log.Fatal(http.ListenAndServe(":3000", nil))

}
