package config

//there is a config in the form of yaml and we need to serialize those values into go code

type HTTPServer struct{
	Addr string `yaml:"address" env:"ADDRESS" env-required:"true"`
}
type Config struct{
	//Capitalizing make those values accessible from other packages
	// package go clean env for easy serializing of the parameters
	Env string `yaml:"env" env:"ENV" env-required:"true"`
	StoragePath string `yaml:"storage-path" env:"STORAGE_PATH" env-required:"true"`
	HTTPServer

}