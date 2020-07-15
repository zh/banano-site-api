package main

import (
	"log"

	"github.com/spf13/viper"
	"go.etcd.io/bbolt"
)

var (
	db     *bbolt.DB
	Config *viper.Viper
)

const accountsBucket = "accounts"

func config() {
	Config = viper.New()
	Config.SetConfigName("banano")
	Config.SetConfigType("json")
	Config.AddConfigPath(".")
	Config.AddConfigPath("./config")
	Config.AddConfigPath("/etc/banano-api/")
    // if not config presents
    Config.SetDefault("AppPort", "8080")
	Config.SetDefault("AppDb", "banano.db")
	Config.SetDefault("AppUser", "api")
	Config.SetDefault("AppPass", "secret")

	err := Config.ReadInConfig()
	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}
}

func main() {
	config()

	db, err := bbolt.Open(Config.Get("AppDb").(string), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, txErr := tx.CreateBucketIfNotExists([]byte(accountsBucket))
		return txErr
	})
	if err != nil {
		log.Fatal(err)
	}

	a := App{}
	a.Initialize(db, Config)
	a.Run(":" + Config.Get("AppPort").(string))

	err = db.Close()
	if err != nil {
		log.Fatal(err)
	}
}
