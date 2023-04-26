package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const StoreFilename = "store.json"

type Entry struct {
	Message string    `json:"message,omitempty"`
	Time    time.Time `json:"time,omitempty"`
	Tag     string    `json:"tag,omitempty"`
}

type Store struct {
	Entries []Entry `json:"entries,omitempty"`
}

type RequestData struct {
	Message string `json:"message,omitempty"`
	Tag     string `json:"tag,omitempty"`
}

func NewStore() *Store {
	return &Store{Entries: []Entry{}}
}

func ReadStore() *Store {

	file, err := os.Open("store.json")
	if err != nil {
		return NewStore()
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var s Store
	err = json.Unmarshal(contents, &s)
	if err != nil {
		return NewStore()
	}

	return &s
}

func WriteStore() error {
	file, err := os.Create("store.json.tmp")
	if err != nil {
		return err
	}

	err = json.NewEncoder(file).Encode(store)
	if err != nil {
		return err
	}

	err = os.Rename("store.json.tmp", "store.json")
	if err != nil {
		return err
	}

	return nil
}

var store *Store

func main() {
	store = ReadStore()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "GET":
			err := json.NewEncoder(w).Encode(store)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				panic(err)
			}
			return

		case "POST":
			var d RequestData
			err := json.NewDecoder(r.Body).Decode(&d)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			time := time.Now()
			store.Entries = append(store.Entries, Entry{Message: d.Message, Time: time, Tag: d.Tag})
			err = WriteStore()
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		case "DELETE":
			store = NewStore()
			err := WriteStore()
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		default:
		}
	})

	http.ListenAndServe(":80", nil)
}
