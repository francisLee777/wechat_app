package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type GlobalConf struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

var Conf GlobalConf
var once sync.Once

func Init() error {
	var err error
	once.Do(func() {
		file, err := os.Open("config.json")
		if err != nil {
			err = fmt.Errorf("failed to open config file: %v", err)
			return
		}
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err = decoder.Decode(&Conf); err != nil || Conf.AppSecret == "" {
			err = fmt.Errorf("failed to decode config file: %v", err)
			panic(err)
			return
		}
	})
	return err
}
