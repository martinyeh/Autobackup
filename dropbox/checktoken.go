package dropbox

import (
	"net/http"
	"io/ioutil"
	"strings"
	"log"
	"github.com/stacktic/dropbox"
)

const errorResp = "The given OAuth 2 access token doesn't exist or has expired."

func CheckToken(token string) bool {

	var db *dropbox.Dropbox
	var err error

	//  Create a new dropbox object.
	db = dropbox.NewDropbox()

	resp, err := http.Get(db.APIURL + "/account/info?access_token=" + token)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	  s := string(body)


	if err != nil {
		log.Println("checktoken error :", err, s)
		return false
	}

	if strings.Contains(s, "error") || strings.Contains(s, errorResp) {
		log.Println(s)
		return false
	}

	return true

}

func GetReAuthToken() string {

	return ""
}
