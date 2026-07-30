package main

//this is the main entry point of the application
import (
	"fmt"
	"log"
	"net/http"
	"pranayteja31/Urlshortener/internal/config"
)

func main() {
	//now the config and its loading has been setup now we are importing it and using it
	//fmt.Println("hello world")

	//loading the congig handle into main
	cfg:= config.MustLoad()

	//setting up the router
	router := http.NewServeMux()
	router.HandleFunc("GET /",func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the url shortener home api"))
	})

	//setup server
	//1. make a new instance and add the server parameters to the server
	Server := &http.Server{
		Addr: cfg.HTTPServer.Addr,
		Handler: router,
	}

	fmt.Println("starting the server")
	
	//start the server with given parameters
	err := Server.ListenAndServe()
	if err != nil {
		log.Fatal("failed to start the server") // error handling at every step
	}

}
