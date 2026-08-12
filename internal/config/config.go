package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// serializing the config paramters
type HTTPServer struct {
	Addr string `yaml:"address" env:"ADDRESS" env-required:"true"` //struct tags are enclosed in `` they help in serialize
}
type DBConfig struct {
	DBHost     string `yaml:"db-host" env:"DB_HOST" env-required:"true"`
	DBPort     string `yaml:"db-port" env:"DB_PORT" env-required:"true"`
	DBUser     string `yaml:"db-user" env:"DB_USER" env-required:"true"`
	DBPassword string `yaml:"db-password" env:"DB_PASSWORD" env-required:"true"`
	DBName     string `yaml:"db-name" env:"DB_NAME" env-required:"true"`
	SSLMode    string `yaml:"db-sslmode" env:"DB_SSLMODE"`
}

// redis congig
type RedisConfig struct {
	Host     string `yaml:"redis-host" env:"REDIS_HOST" env-required:"true"`
	Port     string `yaml:"redis-port" env:"REDIS_PORT" env-required:"true"`
	Password string `yaml:"redis-password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"redis-db" env:"REDIS_DB"`
	PoolSize int    `yaml:"redis-poolsize" env:"REDIS_POOLSIZE"`
}

type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true"`
	HTTPServer  `yaml:"http-server" env-required:"true"`
	DBConfig    `yaml:"db-config" env-required:"true"`
	RedisConfig `yaml:"redis-config"`
}

// loader to load the config parameters
func MustLoad() *Config {
	var configPath string

	configPath = os.Getenv("CONFIG_PATH") //first check with .env files
	if configPath == "" {
		//now if not found through .env
		flags := flag.String("config", "", "path to the congig") // (the name of the flag,the default value, the usage message)
		flag.Parse()
		configPath = *flags //flags returns a pointer to the flags it got and we're storing in the configpath
		//what if the fileds are still missing
		if configPath == "" {
			log.Fatal("No config path found")
		}
	}

	//now let us fetch the data from the path given
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist : %s", configPath) //fatalf is for string formating
	}

	var cfg Config
	err := cleanenv.ReadConfig(configPath, &cfg) //the struct instace cfg is created and the config is read and stored in cfg
	if err != nil {
		log.Fatalf("cannot read the file : %s", err.Error())
	}

	return &cfg
}
