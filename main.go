package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var p = fmt.Printf

func main() {
	rand.Seed(time.Now().UnixNano())
	// parse arguments
	var listId string
	var printJson bool
	var random bool
	for _, arg := range os.Args[1:] {
		if regexp.MustCompile(`^[0-9]+$`).MatchString(arg) {
			listId = arg
		} else if arg == "print_json" {
			printJson = true
		} else if arg == "random" {
			random = true
		}
	}

	// get list json
	if listId == "" {
		p("no list id specified.\n")
		return
	}
	resp, err := http.Get(fmt.Sprintf("http://music.163.com/api/playlist/detail?id=%s&offset=0&total=true&limit=99999", listId))
	if err != nil {
		p("http get error %v.\n", err)
		return
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		p("http read error %v.\n", err)
		return
	}
	resp.Body.Close()

	// print
	if printJson {
		indented := new(bytes.Buffer)
		json.Indent(indented, content, "", "    ")
		p("%s\n", indented)
	}

	// decode
	type Song struct {
		Name string
		Url  string `json:"mp3Url"` // 160k的，320k要麻烦些，反正也听不出区别
	}
	var songs struct {
		Result struct {
			Tracks []Song
		}
		Code int
	}
	err = json.NewDecoder(bytes.NewReader(content)).Decode(&songs)
	if err != nil {
		p("json decode error %v.\n", err)
		return
	}

	// options
	if random {
		tracks := songs.Result.Tracks
		for n := 0; n < 8; n++ {
			for i := 0; i < len(tracks); i++ {
				j := rand.Intn(len(tracks))
				tracks[i], tracks[j] = tracks[j], tracks[i]
			}
		}
	}

	// play
	for _, song := range songs.Result.Tracks {
		p("%s\n", song.Name)
		exec.Command("mpg123", song.Url).Run()
	}
}
