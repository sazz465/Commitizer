package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	numAuthorReviewed = make(map[string]int)
)

func parser(relativeFilePath string) {
	fmt.Println("INSIDE PARSER")
	files, err := ioutil.ReadDir(relativeFilePath)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
		file, err := os.Open(relativeFilePath + f.Name())

		if err != nil {
			log.Fatalf("failed to open")

		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		// var text []string

		for scanner.Scan() {
			ln := strings.TrimSpace(scanner.Text())
			if len(ln) > 12 && ln[:11] == "Reviewed-by" {
				numAuthorReviewed[strings.TrimSpace(ln[12:])]++
			}
		}
		file.Close()
	}

	// for author := range numAuthorReviewed {
	// 	fmt.Printf("\nAuthor %s reviwed %d commits \n", author, numAuthorReviewed[author])
	// }

}
