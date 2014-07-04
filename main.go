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
	var listIds []string
	var printJson bool
	var random bool
	for _, arg := range os.Args[1:] {
		if regexp.MustCompile(`^[0-9]+$`).MatchString(arg) {
			listIds = append(listIds, arg)
		} else if arg == "print_json" {
			printJson = true
		} else if arg == "random" {
			random = true
		}
	}

	type Song struct {
		Name    string
		Url     string `json:"mp3Url"` // 160k的，320k要麻烦些，反正也听不出区别
		Artists []struct {
			Name string
		}
	}
	var songs []Song
	getSongs := func(listId string) {
		// get list json
		if len(listIds) == 0 {
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
		var result struct {
			Result struct {
				Tracks []Song
			}
			Code int
		}
		err = json.NewDecoder(bytes.NewReader(content)).Decode(&result)
		if err != nil {
			p("json decode error %v.\n", err)
			return
		}

		for _, song := range result.Result.Tracks {
			songs = append(songs, song)
		}
	}
	for _, id := range listIds {
		getSongs(id)
	}

	// options
	if random {
		for n := 0; n < 8; n++ {
			for i := 0; i < len(songs); i++ {
				j := rand.Intn(len(songs))
				songs[i], songs[j] = songs[j], songs[i]
			}
		}
	}

	// play
	for _, song := range songs {
		p("%s", song.Name)
		for _, artist := range song.Artists {
			p(" %s", artist.Name)
		}
		p("\n")
		exec.Command("mpg123", song.Url).Run()
	}
}
