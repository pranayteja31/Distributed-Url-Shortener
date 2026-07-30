package main

//this is the main entry point of the application
import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"pranayteja31/Urlshortener/internal/config"
	student "pranayteja31/Urlshortener/internal/http/handlers/students"
	"syscall"
	"time"
)

func main() {
	//now the config and its loading has been setup now we are importing it and using it
	//fmt.Println("hello world")

	//loading the congig handle into main
	cfg:= config.MustLoad()

	//setting up the router
	router := http.NewServeMux()
	router.HandleFunc("GET /api/students",student.New())

	//setup server
	//1. make a new instance and add the server parameters to the server
	Server := &http.Server{
		Addr: cfg.HTTPServer.Addr,
		Handler: router,
	}

	fmt.Println("started server..")
	//enclosing the server to a thread for graeful shutdown
	done := make(chan os.Signal,1)
	signal.Notify(done,os.Interrupt,syscall.SIGINT,syscall.SIGTERM) //if any interrup signals arise the value of channel done changes

	go func ()  {
		//start the server with given parameters
		err := Server.ListenAndServe()
		if err != nil {
			log.Fatal("failed to start the server") // error handling at every step
		}
	}()

	<-done //if the chanel receives any message due to interrupt it means the interrup signal arised

	//Adding graceful shutdown - instead of sudden shutdown it checks and then shuts down the server
	//logic to shutdowm the server because done chan go interrupt signal

	slog.Info("shutting down the server")//logs the info related to shutdown
	ctx,cancel := context.WithTimeout(context.Background(),5*time.Second) //a timeout of 5s
	defer cancel() //cleanup practice of created context
	//defer - runs agter completion of a function

	err := Server.Shutdown(ctx) //shutdown after 5s
	if err != nil {
		slog.Error("failed to shutdown",slog.String("error",err.Error()))
	}
	
	




}
