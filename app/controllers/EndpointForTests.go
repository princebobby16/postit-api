package controllers

import (
	"log"
	"net/http"
)

type p struct {
	Pass string 		`json:"pass"`
}

func EndPointForTests(w http.ResponseWriter, r *http.Request) {
	//var password p
	//err := json.NewDecoder(r.Body).Decode(&password)
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	log.Println(err)
	//	return
	//}
	//log.Println(password)

	log.Println(r.URL)
	if r.URL.Path != "/tests" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}
