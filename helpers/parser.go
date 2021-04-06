package helpers

import (
	"bufio"
	"encoding/csv"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// Parses commit metadata and creates contributions.csv
func Parser(relativeFilePath string, pathCSV string, numAuthorCreated map[string]int, numAuthorReviewed map[string]int) error {
	files, err := ioutil.ReadDir(relativeFilePath)
	if err != nil {
		return err
	}

	err = getReviewerNames(files, relativeFilePath, numAuthorReviewed)
	if err != nil {
		return err
	}

	csvData := make_csv(numAuthorCreated, numAuthorReviewed)

	err = write_csv(relativeFilePath, numAuthorCreated, numAuthorReviewed, csvData)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// Gets all reviewer names from all the previously obtained commit(.txt)
// files of the form commit{index}_hash.txt
func getReviewerNames(files []fs.FileInfo, relativeFilePath string, numAuthorReviewed map[string]int) error {
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
	return nil
}

// Makes CSV file csvData from the maps numAuthorCreated and numAuthorReviewed
func make_csv(numAuthorCreated map[string]int, numAuthorReviewed map[string]int) [][]string {
	csvData := [][]string{{"Contributor", "Created", "Reviewed"}}

	for author := range numAuthorReviewed {
		csvData = append(csvData, []string{author, strconv.Itoa(numAuthorCreated[author]), strconv.Itoa(numAuthorReviewed[author])})
	}

	for author := range numAuthorCreated {
		if _, ok := numAuthorReviewed[author]; !ok {
			csvData = append(csvData, []string{author, strconv.Itoa(numAuthorCreated[author]), strconv.Itoa(numAuthorReviewed[author])})
		}
	}
	return csvData
}

// Writes CSV file csvData into contributions.csv in path relativeFilePath
func write_csv(relativeFilePath string, numAuthorCreated map[string]int, numAuthorReviewed map[string]int, csvData [][]string) error {

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
