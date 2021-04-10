package helpers

import (
	"log"
	"os"
	"strconv"
)

// Makes commit file of the form `commit${commitIndex}_${commitHash}.txt`
func MakeCommitFile(commitMessage string, commitHash string, fpath string, commitIndex int) error {

	path := fpath + "/" + "commit" + strconv.Itoa(commitIndex+1) + "_" + commitHash + ".txt"
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(commitMessage)
	if err2 != nil {
		log.Fatal(err2)
	}
	return err
}
