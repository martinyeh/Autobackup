package dropbox

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/stacktic/dropbox"
	"htc.com/autobackup"
	database "htc.com/db"
)

var (
	chunkSize    = flag.Int("chunksize", 4096*1024, "backup upload chunk size")
	clientId     = flag.String("clientid", "0ln0pu5hc8mojdq", "Client ID of the App")
	clientSecret = flag.String("secret", "lzncsqszaq08rjk", "Client Secret of the App")
	mediaPath    = flag.String("media_path", "/home/martin_yeh/DCIM/", "the folder where stores videos and pictures")
)

func GetToken(path string) string {

	buf := bytes.NewBuffer(nil)
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("cannot open config file", err)
	}
	io.Copy(buf, file)
	defer file.Close()

	var config autobackup.Config

	err = json.Unmarshal(buf.Bytes(), &config)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return config.AccessToken
}

func GenerateFileInfo() {
	id := 0

	db, _ := database.CreateTable("./fileinfo.db")

	files, _ := ioutil.ReadDir(*mediaPath)

	for _, f := range files {
		err := database.InsertTable(db, id, f.Name(), false)
		if err != nil {
			fmt.Println(err)
			continue
		}
		id++
	}
}

func isComplete(db *database.DB, fn string) bool {
	fmt.Println("isComplete", fn)

	rows, err := db.QueryTable(fn)

	if err != nil {
		fmt.Println("query table info failure")
		return false
	}

	for rows.Next() {
		var id int
		var fn string
		var complete int
		rows.Scan(&id, &fn, &complete)
		fmt.Println(id, fn, complete)
		if complete == 1 {
			return true
		}
	}
	rows.Close()

	return false

}

func BatchUpload(token string) {


	if(CheckToken(token)== false){
		return
	}


	var db *dropbox.Dropbox
	var err error

	//  Create a new dropbox object.
	db = dropbox.NewDropbox()

	db.SetAppInfo(*clientId, *clientSecret)

	// Provide the user token.
	db.SetAccessToken(token)
	token = db.AccessToken()
	log.Println("oauth2 access token: ", token)

	// Send your commands.
	//  you will create a new folder named "GC".
	folder := "RE/"
	if _, err = db.CreateFolder(folder); err != nil {
		fmt.Printf("Error creating folder %s: %s\n", folder, err)
	} else {
		fmt.Printf("Folder %s successfully created\n", folder)
	}

	var fd *os.File

	files, _ := ioutil.ReadDir(*mediaPath)

	//	dtbs := database.NewConn("./fileinfo.db")

	for _, f := range files {
		dtbs := database.NewConn("./fileinfo.db")
		fmt.Println(*mediaPath, f.Name())

		if isComplete(dtbs, f.Name()) {
			continue
		}

		if fd, err = os.Open(*mediaPath + f.Name()); err != nil {
			log.Println("open file error")
			return
		}
		defer fd.Close()

		overwrite := false
		parentRev := ""
		dst := folder + f.Name()

		entry, err := UploadByChunk(db, fd, *chunkSize, dst, overwrite, parentRev)
		if err != nil {
			log.Println("upload failure:", err)
			continue
		}
		log.Println(entry)

		fmt.Println("update Table....")
		dtbs = database.NewConn("./fileinfo.db")
		dtbs.UpdateTable(f.Name(), true)
	}

}

func writeToFile(cur *dropbox.ChunkUploadResponse, path string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Printf("Warning: failed to cache config: %v", err)
		return err
	}
	defer f.Close()

	data, err := json.Marshal(cur)
	if err != nil {
		log.Printf("Warning: failed to cache config: %v", err)
		return err
	}

	f.Write(data)
	return nil
}

func readFromFile(path string) (*dropbox.ChunkUploadResponse, error) {
	buf := bytes.NewBuffer(nil)
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("cannot open config file", err)
		return nil, err
	}
	io.Copy(buf, file)
	defer file.Close()

	var cur *dropbox.ChunkUploadResponse

	err = json.Unmarshal(buf.Bytes(), &cur)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return cur, nil

}

// UploadByChunk uploads data from the input reader to the dst path on Dropbox by sending chunks of chunksize.
func UploadByChunk(db *dropbox.Dropbox, fd *os.File, chunksize int, dst string, overwrite bool, parentRev string) (*dropbox.Entry, error) {
	var err error
	var cur *dropbox.ChunkUploadResponse

	filename := "resp.json"

	for err == nil {
		if _, err := os.Stat(filename); err == nil {
			cur, err = readFromFile(filename)
			ret, _ := fd.Seek(int64(cur.Offset), 0)
			log.Println("current offset:", ret)
		}

		if cur, err = db.ChunkedUpload(cur, fd, chunksize); err != nil && err != io.EOF {
			if cur != nil {
				ret, _ := fd.Seek(int64(cur.Offset), 0)
				log.Println("====================Resend upload request=================")
				log.Println("resend current offset:", ret)
				cur, err = db.ChunkedUpload(cur, fd, chunksize)
			}

			if err != nil && err != io.EOF {
				return nil, err
			}
		}
		writeToFile(cur, filename)

	}

	entry, err := db.CommitChunkedUpload(cur.UploadID, dst, overwrite, parentRev)

	os.Remove(filename)

	return entry, err
}

/*func main() {
	token := GetToken(*autobackup.ConfigPath)
	fmt.Println(token)

	var db *dropbox.Dropbox

	// 1. Create a new dropbox object.
	db = dropbox.NewDropbox()

	// 2. Provide your clientid and clientsecret (see prerequisite).
	//db.SetAppInfo(clientid, clientsecret)

	// 3. Provide the user token.
	// This method will ask the user to visit an URL and paste the generated code.
	//if err = db.Auth(); err != nil {
	//   fmt.Println(err)
	//  return
	//}
	// You can now retrieve the token if you want.
	db.SetAccessToken(token)
	token = db.AccessToken()
	fmt.Println("token: " + token)

	// 4. Send your commands.
	// In this example, you will create a new folder named "demo".
	//folder := "demo"
	//if _, err = db.CreateFolder(folder); err != nil {
	//   fmt.Printf("Error creating folder %s: %s\n", folder, err)
	//} else {
	//   fmt.Printf("Folder %s successfully created\n", folder)
	//}

	//Create a upload session
	//var session *dropbox.ChunkUploadResponse
	//var resp *dropbox.ChunkUploadResponse
	var fd *os.File
	//session = nil

	path := "/home/martin_yeh/Videos/video2.3gp"
	if fd, err = os.Open(path); err != nil {
		fmt.Println("open file error", err)
		return
	}
	defer fd.Close()

	overwrite := true
	parentRev := ""
	chunksize := 1024 * 256
	dst := "sample/a.3gp"

	entry, err := db.UploadByChunk(fd, chunksize, dst, overwrite, parentRev)
	if err != nil {
		fmt.Println("upload error:", err)
	}
	fmt.Println(entry)
}*/
