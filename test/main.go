package main

import (
	"flag"
	"fmt"
	"htc.com/autobackup"
	"htc.com/dropbox"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {

	flag.Parse()

	accessToken := dropbox.GetToken(*autobackup.ConfigPath)

	fmt.Println(accessToken)

	//sendBackupConfig(accessToken)

	if _, err := os.Stat("fileinfo.db"); os.IsNotExist(err) {
		dropbox.GenerateFileInfo()
	}

	batchUpload(accessToken)

}

func sendBackupConfig(accessToken string) {

	resp, err := http.PostForm("http://localhost:3000/cloud/backup_worker/settings",
		url.Values{"provider": {"dropbox"}, "access_token": {accessToken}, "refresh_token": {""}, "iskeep": {""}})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)

}

func batchUpload(token string) {

	start := time.Now()
	dropbox.BatchUpload(token)

	elapsed := time.Since(start)
	fmt.Println("Upload time :", elapsed)

}
