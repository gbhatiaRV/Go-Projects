package main

import (
	"fmt"
	"unicode"
)

func main() {
	var str string
	fmt.Println("Please Enter the string")
	fmt.Scanln(&str)
	//var word []string
	w := ""
	index := 1
	firstRun := true

	//fmt.Println(str + string(len(str)))

	for _, s := range str {
		if unicode.IsUpper(s) {

			if !firstRun {
				//fmt.Println(string(s))
				//println(w)
				fmt.Printf(w + " ")
				//println(index)
				//word[index] = w
				index = index + 1
				w = ""
			}

		}
		w = w + string(s)
		firstRun = false
	}
	fmt.Println(w)
	fmt.Println("String Contains ", index, " words")
	//fmt.Println(word[0])
}
