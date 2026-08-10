package main

//this is the main entry point of the application
import (
	"pranayteja31/Urlshortener/internal/cache"
	"pranayteja31/Urlshortener/internal/config"
	"pranayteja31/Urlshortener/internal/db"
	"pranayteja31/Urlshortener/internal/handlers"
	middleware "pranayteja31/Urlshortener/internal/middlerware"
	"pranayteja31/Urlshortener/internal/repository"
	"pranayteja31/Urlshortener/internal/routes"
	"pranayteja31/Urlshortener/internal/services"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pranayteja31/Urlshortener/internal/metrics"
)



func main() {
	cfg := config.MustLoad()

	dbConn := db.DB_connection()
	defer dbConn.Close()
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
	services := services.NewURLServices(repo,cacheStore)
	//register handlers with the services
	handlers := handlers.NewHandler(services)

	//registering the routers with handlers
	routes.RegisterRoutes(router, handlers,redisClient)

	//start the server
	err = router.Run(cfg.HTTPServer.Addr)
	if err != nil {
		panic(err)
	}
}
