package main

//this is the main entry point of the application
import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"pranayteja31/Urlshortener/internal/analytics"
	"pranayteja31/Urlshortener/internal/cache"
	"pranayteja31/Urlshortener/internal/config"
	"pranayteja31/Urlshortener/internal/db"
	"pranayteja31/Urlshortener/internal/handlers"
	middleware "pranayteja31/Urlshortener/internal/middlerware"
	"pranayteja31/Urlshortener/internal/repository"
	"pranayteja31/Urlshortener/internal/routes"
	"pranayteja31/Urlshortener/internal/services"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pranayteja31/Urlshortener/internal/metrics"
)



func main() {
	cfg := config.MustLoad()

	dbConn := db.DB_connection()
	defer dbConn.Close()
	//analytics
	clickRepo := repository.NewClickRepository(dbConn)
	analyticsWorker := analytics.NewWorker(clickRepo,100)

	Workerctx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	go analyticsWorker.Start(Workerctx)
	
	//init Redis
	redisClient,err := cache.NewRedisClient(&cfg.RedisConfig)
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	cacheStore := cache.NewRedisCache(redisClient)

	metrics.Register()
	//init the router
	router := gin.Default()

	//metrics
	router.Use(middleware.PrometheusMiddleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	//register repos with db
	repo := repository.NewRepository(dbConn)
	//register services with repos
	services := services.NewURLServices(repo,cacheStore,analyticsWorker)
	//register handlers with the services
	handlers := handlers.NewHandler(services)

	//registering the routers with handlers
	routes.RegisterRoutes(router, handlers,redisClient)

	//server
	server := &http.Server{
		Addr: cfg.HTTPServer.Addr,
		Handler: router,
	}

	//listen for the interrupts
	stop := make(chan os.Signal,1)
	signal.Notify(stop,os.Interrupt,syscall.SIGTERM)
	defer signal.Stop(stop)
	//start the http server
	go func ()  {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err,http.ErrServerClosed) {
			panic(err)
		}
	}()
	//waiting for shutdown
	<-stop
	//stopping the server
	shutdownCtx,cancelShutdown := context.WithTimeout(context.Background(),10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx);err != nil {
		panic(err)
	}

	cancelWorker()
	analyticsWorker.Wait()

}
