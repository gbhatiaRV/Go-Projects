package main

import (
	//"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	// Open the file
	var fileName, userFileName string
	fmt.Println("Enter CSV FileName")
	fmt.Scanln(&userFileName) // If user wants to give the fileName
	

	if fileName == "" {
		fileName = "problems.csv"
	}

	csvfile, err := os.Open(fileName)
	var userAnswer string

	if err != nil {
		log.Fatalln("cannot open the file. Please check your File Name", err)
	}

	// Parse the file
	r := csv.NewReader(csvfile)

	Score := 0
	total := 0
	// Iterate through the records
	for {
		// Read each record from csv
		//println(Score)
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		total++

		fmt.Printf("Question: %s \n", record[0])
		fmt.Scanln(&userAnswer)
		if userAnswer == record[1] {
			userAnswer = ""
			//println(Score)
			Score++
		}

	}

	fmt.Println("Your total score is", Score, "out of", total)
}
