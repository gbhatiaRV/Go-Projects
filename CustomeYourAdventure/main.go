package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

//Options to read story optuions
type Options struct {
	Text string `json:"text"`
	Arc  string `json:"Arc"`
}

//Chapter to read story details
type Chapter struct {
	Title     string   `json:"title"`
	ParaStory []string `json:"Story"`
	Options   []Options
}

func init() {
	tpl = template.Must(template.ParseFiles("story.html"))
}

var tpl *template.Template

//Story is chapter variable
type Story map[string]Chapter

var story Story

//function process the story data and put it in story.html template
func storyHandler(s Story) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimLeft(r.URL.Path, "/")
		tm := s[path]
		if path == "" {
			tm = s["intro"]
		}

		tpl.ExecuteTemplate(w, "story.html", tm)
	}
	return http.HandlerFunc(fn)
}

func clickLink(arc string, s Story) {
	shandler := storyHandler(s)
	http.Handle(arc, shandler)
}

//HandleJSON bla bla bla
func main() {
	// Open our jsonFile
	jsonFile, err := os.Open("gophers.json")

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	//Converting JSON to story which is of type Story a chapter variable
	json.Unmarshal(byteValue, &story)

	if err != nil {
		fmt.Println(err)
	}

	clickLink("/", story) // function takes story path and story as parameters and call a handler

	log.Println("Listening...")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
