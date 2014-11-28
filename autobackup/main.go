package autobackup

import (
	"bytes"
	"code.google.com/p/goauth2/oauth"
	"encoding/gob"
        "net/textproto"
        "mime/multipart"
	"errors"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
        //"io/ioutil"
)

var config = &oauth.Config{
	ClientId:     "1086179271861-2tsvi5faqcb2tn3mh85dmr19vhsa4v4t.apps.googleusercontent.com", // Set by --clientid or --clientid_file
	ClientSecret: "gygqPrsBEdlO3MWODcGDiBfw",                                                  // Set by --secret or --secret_file
	RedirectURL:  *redirectURL,
	Scope:        "https://picasaweb.google.com/data/", // filled in per-API
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
	TokenCache:   oauth.CacheFile(*cachefile),
	AccessType:   "offline",
}

// Flags
var (
	//clientId     = flag.String("clientid", "1086179271861.project.googleusercontent.com", "OAuth Client ID.  If non-empty, overrides --clientid_file")
	//clientSecret = flag.String("secret", "kcZbD1Xy928krFu6WvHdOFsM", "Client Secret")
	redirectURL  = flag.String("redirect_url", "urn:ietf:wg:oauth:2.0:oob", "Redirect URL")
	clientIdFile = flag.String("clientid_file", "clientid.dat",
		"Name of a file containing just the project's OAuth Client ID from https://code.google.com/apis/console/")
	secretFile = flag.String("secret_file", "clientsecret.dat",
		"Name of a file containing just the project's OAuth Client Secret from https://code.google.com/apis/console/")
	cacheToken = flag.Bool("cachetoken", true, "cache the OAuth token")
	code       = flag.String("code", "", "Authorization Code")
	requestURL = flag.String("request_url", "https://picasaweb.google.com/data/feed/api/user/default/albumid/6075534675049165505", "API request")
	debug      = flag.Bool("debug", false, "show HTTP traffic")
	cachefile  = flag.String("cache", "cache.json", "Token cache file")
)

const usageMsg = `
To obtain a request token you must specify both -id and -secret.

To obtain Client ID and Secret, see the "OAuth 2 Credentials" section under
the "API Access" tab on this page: https://code.google.com/apis/console/

Once you have completed the OAuth flow, the credentials should be stored inside
the file specified by -cache and you may run without the -id and -secret flags.
`

// Creates a new file upload http request
func newImageUploadRequest(uri string, path string) (*http.Request, error) {
	file, err := os.Open(path)
	fi, err := file.Stat()
	if err != nil {
		// Could not obtain stat, handle error
	}

	body := &bytes.Buffer{}
	length := strconv.Itoa(int(fi.Size()))

	io.Copy(os.Stdout, body)
	fmt.Println()

	if err == nil {
		io.Copy(body, file)
	}
	defer file.Close()

	req, err := http.NewRequest("POST", *requestURL, body)

	req.Header.Add("GData-Version", "2")
	req.Header.Add("Content-Type", "image/jpeg")
	req.Header.Add("Content-Length", length)

	return req, err
}

func newVideoUploadRequest(uri string, path string) (*http.Request, error) {
        file, err := os.Open(path)
        fi, err := file.Stat()
        if err != nil {
                // Could not obtain stat, handle error
        }

        body := &bytes.Buffer{}
        length := strconv.Itoa(int(fi.Size()))
        fmt.Println(length)

        body_writer := multipart.NewWriter(body)
        boundary := body_writer.Boundary()
	fmt.Println(boundary + "$$$$$$$$$$\n")

        mh := make(textproto.MIMEHeader)
        mh.Set("Content-Type", "application/atom+xml")
        part_writer, err := body_writer.CreatePart(mh)
        if nil != err {
           panic(err.Error())
        }
	io.Copy(part_writer, bytes.NewBufferString("<entry xmlns='http://www.w3.org/2005/Atom'> \n" + 
        "<title>video.wmv</title> \n"+
        "<summary>Real cat wants attention too.</summary> \n"+
        "<category scheme=\"http://schemas.google.com/g/2005#kind\" \n"+
        "term=\"http://schemas.google.com/photos/2007#photo\"/> \n"+
        "</entry>\n\n--" + boundary ))

	body.WriteString("\nContent-type: video/x-ms-wmv\n") 

	if err == nil {
	       io.Copy(body, file)
	}
	defer file.Close()

	body.WriteString("\n--"+ boundary+"--")

        req, err := http.NewRequest("POST", *requestURL, body)

        req.Header.Add("GData-Version", "2")
        req.Header.Add("Content-Length", length)
        req.Header.Add("Content-Type", "multipart/related; boundary=\"" + boundary+"\"")

        return req, err
}



func main() {
	flag.Parse()

	client := getOAuthClient(config)
	//var path = "/home/martin_yeh/Pictures/app_logo.jpg"
        var path = "/home/martin_yeh/Videos/Justdrive2.wmv" 

	req, err := newVideoUploadRequest(*requestURL, path)
     
	if err == nil {
		r, err := client.Do(req)

		if err == nil {
			// Write the response to standard output.
			io.Copy(os.Stdout, r.Body)

			// Send final carriage return, just to be neat.
			fmt.Println()
		}
	}

}

func getOAuthClient(config *oauth.Config) *http.Client {
	cacheFile := tokenCacheFile(config)
	token, err := tokenFromFile(cacheFile)
	if err != nil {
		token = tokenFromWeb(config)
		saveToken(cacheFile, token)
	} else {
		log.Printf("Using cached token %#v from %q", token, cacheFile)
	}

	t := &oauth.Transport{
		Token:     token,
		Config:    config,
		Transport: condDebugTransport(http.DefaultTransport),
	}
	return t.Client()
}

func tokenFromWeb(config *oauth.Config) *oauth.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	authUrl := config.AuthCodeURL(randState)
	go openUrl(authUrl)
	log.Printf("Authorize this app at: %s", authUrl)
	code := <-ch
	log.Printf("Got code: %s", code)

	t := &oauth.Transport{
		Config:    config,
		Transport: condDebugTransport(http.DefaultTransport),
	}
	_, err := t.Exchange(code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}
	return t.Token
}

func osUserCacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache")
	}
	log.Printf("TODO: osUserCacheDir on GOOS %q", runtime.GOOS)
	return "."
}

func tokenCacheFile(config *oauth.Config) string {
	hash := fnv.New32a()
	hash.Write([]byte(config.ClientId))
	hash.Write([]byte(config.ClientSecret))
	hash.Write([]byte(config.Scope))
	fn := fmt.Sprintf("go-api-demo-tok%v", hash.Sum32())
	return filepath.Join(osUserCacheDir(), url.QueryEscape(fn))
}

func tokenFromFile(file string) (*oauth.Token, error) {
	if !*cacheToken {
		return nil, errors.New("--cachetoken is false")
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := new(oauth.Token)
	err = gob.NewDecoder(f).Decode(t)
	return t, err
}

func saveToken(file string, token *oauth.Token) {
	f, err := os.Create(file)
	if err != nil {
		log.Printf("Warning: failed to cache oauth token: %v", err)
		return
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(token)
}

func condDebugTransport(rt http.RoundTripper) http.RoundTripper {
	if *debug {
		return &logTransport{rt}
	}
	return rt
}

func openUrl(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Printf("Error opening URL in browser.")
}
