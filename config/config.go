package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	Env     = os.Getenv("GO_ENV")
	AppRoot = filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "albrow", "5w4g-server")
)

var (
	Secret         []byte
	Host           string
	Port           string
	AllowOrigins   []string
	PrivateKey     []byte
	PrivateKeyFile string
	Aws            awsConfig
	Db             dbConfig
)

type config struct {
	Secret         []byte
	Host           string
	Port           string
	AllowOrigins   []string
	PrivateKeyFile string
	Db             dbConfig
	Aws            awsConfig
}

type dbConfig struct {
	Address  string
	Network  string
	Database int
}

type awsConfig struct {
	AccessKeyId     string
	SecretAccessKey string
}

var Prod config = config{
	Secret:         []byte(""), // TODO: Set secret based on environment variable
	Host:           "",         // TODO: Set this to our api domain
	Port:           "8080",
	AllowOrigins:   []string{"5w4g.com", "admin.5w4g.com"}, // TODO: Only allow requests from our static content domain
	PrivateKeyFile: filepath.Join(AppRoot, "config", "id.rsa"),
	Db: dbConfig{
		Address:  "", // TODO: Set to our redis server domain
		Network:  "tcp",
		Database: 0,
	},
	Aws: awsConfig{
		AccessKeyId:     os.Getenv("SWAG_AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("SWAG_AWS_SECRET_ACCESS_KEY"),
	},
}

var Dev config = config{
	Secret:         []byte("5776cb45330d129c6f28e82ff4d9868ac9917def536dfa9f5680bf1c6a8a3f2e"),
	Host:           "localhost",
	Port:           "3000",
	AllowOrigins:   []string{"*"},
	PrivateKeyFile: filepath.Join(AppRoot, "config", "dev_id.rsa"),
	Db: dbConfig{
		Address:  "localhost:6379",
		Network:  "tcp",
		Database: 11,
	},
	Aws: awsConfig{
		AccessKeyId:     os.Getenv("SWAG_AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("SWAG_AWS_SECRET_ACCESS_KEY"),
	},
}

var Test config = config{
	Secret:         []byte("1bc217e1e32d91b1769ae3d15a0f2f65138b84e51bf63f574707eb56304023fb"),
	Host:           "localhost",
	Port:           "4000",
	AllowOrigins:   []string{"*"},
	PrivateKeyFile: filepath.Join(AppRoot, "config", "test_id.rsa"),
	Db: dbConfig{
		Address:  "localhost:6379",
		Network:  "tcp",
		Database: 12,
	},
	Aws: awsConfig{
		AccessKeyId:     os.Getenv("SWAG_AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("SWAG_AWS_SECRET_ACCESS_KEY"),
	},
}

func Init() {
	requireEnvVariables("SWAG_AWS_ACCESS_KEY_ID", "SWAG_AWS_SECRET_ACCESS_KEY")
	if Env == "development" || Env == "" {
		Env = "development"
		Use(Dev)
	} else if Env == "test" {
		Use(Test)
	} else if Env == "production" {
		Use(Prod)
	} else {
		panic("Unkown environment. Don't know what configuration to use!")
	}
	readPrivateKey()
	fmt.Printf("[config] Running in %s environment...\n", Env)
}

func readPrivateKey() {
	if key, err := ioutil.ReadFile(PrivateKeyFile); err != nil {
		panic(err)
	} else {
		PrivateKey = key
	}
}

func requireEnvVariables(keys ...string) {
	for _, key := range keys {
		if os.Getenv(key) == "" {
			panic("Missing required environment variable: " + key)
		}
	}
}

func Use(c config) {
	Secret = c.Secret
	Host = c.Host
	Port = c.Port
	AllowOrigins = c.AllowOrigins
	PrivateKeyFile = c.PrivateKeyFile
	Db = c.Db
	Aws = c.Aws
}
