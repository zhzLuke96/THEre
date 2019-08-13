package main

import (
	"flag"
	"net/http"
	"time"

	Rango "./Rango"
	RangoKit "./Rango/core"
	"./Rango/mid"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "8080", "server port.")
	flag.StringVar(&port, "p", "8080", "server port.")

	flag.Parse()
	MainSev(port)
	return
}

func clearCookie(w RangoKit.ResponseWriteBody, r *http.Request) {
	c := new(http.Cookie)
	c.Name = "RangoToken"
	c.Expires = time.Now().AddDate(-1, 0, 0)
	http.SetCookie(w, c)
	oc, _ := r.Cookie("RangoToken")
	oc.Expires = time.Now().AddDate(-1, 0, 0)
}

func MainSev(port string) {
	sev := Rango.NewSev()
	router := Rango.NewRouter()

	sev.Use(mid.LogRequest)
	sev.Use(mid.ErrCatch)
	sev.Use(router.Mid)

	fileRouter := Rango.NewRouter()
	fileHandler := Rango.NewSev()

	fileHandler.Use(fileRouter.Mid)

	// 路由设置
	router.RangoFunc("/doc/{index:[0-9]+}", getDoc).Methods("GET")
	router.RangoFunc("/docs/{pageNum:[0-9]+}", getDocs).Methods("GET")
	router.RangoFunc("/addDoc", addDoc).Methods("POST")
	router.RangoFunc("/delDoc/{index:[0-9]+}", delDoc).Methods("GET")
	router.RangoFunc("/delDocTitle/{title:.+}", delDocTitle).Methods("GET")

	router.Handler("/", fileHandler)

	fileRouter.Handler("/", http.StripPrefix("/", http.FileServer(http.Dir("./www"))))

	sev.Go(port)
}

// func htmlKeywordMid(w RangoKit.ResponseWriteBody, req *http.Request, next func()) {

// 	next()
// }
