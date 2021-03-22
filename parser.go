package main

import (
	"bufio"
	"encoding/csv"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func parser(relativeFilePath string, pathCSV string) error {
	files, err := ioutil.ReadDir(relativeFilePath)
	if err != nil {
		return err
	}
	for _, f := range files {
		file, err := os.Open(relativeFilePath + f.Name())

		if err != nil {
			return err

		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			ln := strings.TrimSpace(scanner.Text())
			if len(ln) > 12 && ln[:11] == "Reviewed-by" {
				numAuthorReviewed[strings.TrimSpace(ln[12:])]++
			}
		}
		file.Close()
	}

	csvData := [][]string{{"Author", "Created", "Reviewed"}}
	// var authorList []string

	for author := range numAuthorReviewed {
		// authorList = append(authorList, author)
		csvData = append(csvData, []string{author, strconv.Itoa(numAuthorCreated[author]), strconv.Itoa(numAuthorReviewed[author])})
		// fmt.Printf("\nAuthor %s reviwed %d commits and created %d  \n", author, numAuthorReviewed[author], numAuthorCreated[author])
	}

	for author := range numAuthorCreated {
		if _, ok := numAuthorReviewed[author]; !ok {
			// authorList = append(authorList, strconv.Itoa(val))
			csvData = append(csvData, []string{author, strconv.Itoa(numAuthorCreated[author]), strconv.Itoa(numAuthorReviewed[author])})
		}
		// fmt.Printf("\nAuthor %s reviwed %d commits and created %d  \n", author, numAuthorReviewed[author], numAuthorCreated[author])
	}

	err = make_csv(relativeFilePath, numAuthorCreated, numAuthorReviewed, csvData)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func make_csv(relativeFilePath string, numAuthorCreated map[string]int, numAuthorReviewed map[string]int, csvData [][]string) error {

	file, err := os.Create(relativeFilePath + "/" + "contributions.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range csvData {
		err = writer.Write(value)
		if err != nil {
			return err
		}
	}

	return nil
}
