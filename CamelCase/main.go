package main

import (
	"fmt"
	"unicode"
)

func main() {
	var str string
	fmt.Println("Please Enter the string") //Taking Camel Case String from User
	fmt.Scanln(&str)

	w := ""
	numWords := 1
	firstRun := true

	for _, s := range str {
		if unicode.IsUpper(s) {
			if !firstRun {
				fmt.Printf(w + " ")
				numWords = numWords + 1
				w = ""
			}

		}
		w = w + string(s)
		firstRun = false
	}
	fmt.Println(w)
	fmt.Println("String Contains ", numWords, " words")
}
