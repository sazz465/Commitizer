package helpers

import (
	"bufio"
	"encoding/csv"
	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Parses commit metadata and creates contributions.csv
func Parser(commitsFilePath string, csvPath string, numAuthorCreated map[string]int, numAuthorReviewed map[string]int) error {
	files, err := ioutil.ReadDir(commitsFilePath)
	if err != nil {
		return errors.Wrap(err, "couldn't read path directory passed in commitsFilePath")
	}

	err = getReviewerNames(files, commitsFilePath, numAuthorReviewed)
	if err != nil {
		return err
	}

	csvData := makeCSV(numAuthorCreated, numAuthorReviewed)

	err = writeCSV(csvPath, numAuthorCreated, numAuthorReviewed, csvData)
	if err != nil {
		return err
	}
	return nil
}

// Gets all reviewer names from all the previously obtained commit(.txt)
// files of the form commit{index}_hash.txt
func getReviewerNames(files []fs.FileInfo, commitsFilePath string, numAuthorReviewed map[string]int) error {
	for _, f := range files {
		file, err := os.Open(commitsFilePath + "/" + f.Name())
		if err != nil {
			return errors.Wrap(err, "couldn't open file with path directory passed in commitsFilePath")
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
func makeCSV(numAuthorCreated map[string]int, numAuthorReviewed map[string]int) [][]string {
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

// Writes CSV file csvData into contributions.csv in path csvFilePath
func writeCSV(csvFilePath string, numAuthorCreated map[string]int, numAuthorReviewed map[string]int, csvData [][]string) error {

	file, err := os.Create(csvFilePath + "/" + "contributions.csv")
	if err != nil {
		return errors.Wrap(err, "couldn't create contributions.csv")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range csvData {
		err = writer.Write(value)
		if err != nil {
			return errors.Wrap(err, "couldn't write data to csvData")
		}
	}

	return nil
}
