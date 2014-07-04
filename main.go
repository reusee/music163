package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
		} else if arg == "nao" {
			listIds = []string{"20219787"}
		} else if arg == "anisong" {
			listIds = []string{"18687814", "18689435", "18678737", "18474223", "18388961", "18391086", "18389611"}
		}
	}

	type Song struct {
		Id      int
		Name    string
		Artists []struct {
			Name string
		}
		Mp3 struct {
			DfsId uint64
		} `json:"bMusic"`
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
	enc := func(id uint64) string {
		b1 := []byte("3go8&$8*3*3h0k(2)2")
		b2 := []byte(fmt.Sprintf("%d", id))
		b1Len := len(b1)
		for i, b := range b2 {
			b2[i] = b ^ b1[i%b1Len]
		}
		h := md5.New()
		h.Write(b2)
		res := base64.StdEncoding.EncodeToString(h.Sum(nil))
		res = strings.Replace(res, "/", "_", -1)
		res = strings.Replace(res, "+", "-", -1)
		return res
	}
	for _, song := range songs {
		p("%s", song.Name)
		for _, artist := range song.Artists {
			p(" %s", artist.Name)
		}
		url := fmt.Sprintf("http://m2.music.126.net/%s/%d.mp3", enc(song.Mp3.DfsId), song.Mp3.DfsId)
		p(" http://music.163.com/#/song?id=%d", song.Id)
		p("\n")
		exec.Command("mpg123", url).Run()
	}
}
