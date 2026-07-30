package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

//there is a config in the form of yaml and we need to serialize those values into go code

type HTTPServer struct{
	Addr string `yaml:"address" env:"ADDRESS" env-required:"true"` //struct tags are enclosed in `` they help in serialize
}
type Config struct{
	//Capitalizing make those values accessible from other packages
	// package go clean env for easy serializing of the parameters
	Env string `yaml:"env" env:"ENV" env-required:"true"`
	StoragePath string `yaml:"storage-path" env:"STORAGE_PATH" env-required:"true"`
	HTTPServer

}
 
//the struct is added to store the config variable and the serializzation is added
// now we need to fetch the values and store them and load our server
// for which we are using the mustload func

func MustLoad() *Config {
	//now first we have a yaml file with all those necessay parameters 
	// while running the server we want the path of that yaml file so that we can fetch
	var configPath string
	//we are taking the i/p of config path from the flags of terminal
	//flags are --> go build dlfj/dkhfsd -u .....(the things u mention here are called flags)

	configPath = os.Getenv("CONFIG_PATH") //first check with .env files
	if configPath == "" {
		//now if not found through .env
		flags := flag.String("config","","path to the congig") // (the name of the flag,the default value, the usage message)
		flag.Parse()
		configPath = *flags //flags returns a pointer to the flags it got and we're storing in the configpath
		//what if the fileds are still missing
		if configPath == "" {
			log.Fatal("No config path found")
		}
	}

	//now let us fetch the data from the path given
	if _,err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist : %s",configPath) //fatalf is for string formating
	}

	var cfg Config
	err := cleanenv.ReadConfig(configPath,&cfg) //the struct instace cfg is created and the config is read and stored in cfg
	if err != nil {
		log.Fatalf("cannot read the file : %s",err.Error())
	}

	return &cfg
}