package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"flag"

	"github.com/go-martini/martini"
	"htc.com/autobackup/wifimgr"
	"htc.com/autobackup"
)


var configPath *string

func init() {
	  configPath  = flag.String("configPath", "../dropbox/backup_config.json", "backup tokens and some settings")
}


func main() {
        flag.Parse()

	m := martini.Classic()
	m.Get("/", func() string {
		return "OK"
	})

	m.Post("/cloud/wifimgr/settings", wifiSettings)
	m.Post("/cloud/backup_worker/settings", backupSettings)

	m.Run()
}

func wifiSettings(req *http.Request) (int, string) {
	var response string

	response = "POST\n\n"

	// if we would use this, then the POST and GET requests would be merged
	// to the req.Form variable
	// req.ParseForm()
	// v := req.Form

	body, _ := ioutil.ReadAll(req.Body)
	v, _ := url.ParseQuery(string(body))

	config := wifimgr.Config{}

	for key, value := range v {
		fmt.Printf("key: %v\n", key)
		fmt.Printf("value: %v\n", value)

		if key == "ssid" {
			config.Ssid = value[0]
		} else if key == "password" {
			config.Password = value[0]
		} else if key == "security" {
			config.Security = value[0]
		}

	}
	result, _ := json.Marshal(config)
	fmt.Println(string(result))


	err := saveConfig("./gcwifi.json", result)
	if err != nil {
		return 404, "can not create file"
	}

	response += "{'status', 'ok'}"

	return 200, response

}

func backupSettings(req *http.Request) (int, string) {
	var response string

	response = "POST\n\n"

	body, _ := ioutil.ReadAll(req.Body)
	v, _ := url.ParseQuery(string(body))

	log.Printf( "query params: ", v)

	config := autobackup.Config{}

	for key, value := range v {
		fmt.Printf("key: %v\n", key)
		fmt.Printf("value: %v\n", value)

		if key == "provider" {
			config.Provider = value[0]
		} else if key == "access_token" {
			config.AccessToken = value[0]
		} else if key == "refresh_token" {
			config.RefreshToken = value[0]
		} else if key == "iskeep" {
			config.Iskeep,_ = strconv.ParseBool(value[0])
		}

	}
	result, _ := json.Marshal(config)
	fmt.Println(string(result))

	err := saveConfig(*configPath, result)
	if err != nil {
		return 404, "can not create file"
	}

	response += "{'status', 'saved'}"

	return 200, response

}

func saveConfig(file string, json []byte) error{
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache config: %v", err)
		return err
	}
	defer f.Close()
	f.Write(json)
	return nil
}
