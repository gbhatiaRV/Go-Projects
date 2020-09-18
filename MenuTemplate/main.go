upackage main

import (
	"html/template"
	"log"
	"os"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseFiles("menu.html"))
}

type dish struct {
	item string
}

type menu struct {
	dishes   []dish
	menuType string
}

func main() {

	var m []menu
	var d, d1 []dish

	d = []dish{dish{"Eggs"}, dish{"toast"}, dish{"hashBrown"}}
	d1 = []dish{dish{"Pizza"}, dish{"Pasta"}, dish{"Salad"}, dish{"Tomato Soup"}}

	m = []menu{menu{
		dishes:   d,
		menuType: "BreakFast",
	},
		menu{
			dishes:   d1,
			menuType: "Lunch",
		}}

	err := tpl.Execute(os.Stdout, m)
	if err != nil {
		log.Fatalln(err)
	}

}
