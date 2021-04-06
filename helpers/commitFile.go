package helpers

import (
	"log"
	"os"
	"strconv"
)

// Makes commit file of the form `commit${commitIndex}_${commit_hash}.txt`
func MakeCommitFile(commit_message string, commit_hash string, fpath string, commitIndex int) error {

	path := fpath + "/" + "commit" + strconv.Itoa(commitIndex+1) + "_" + commit_hash + ".txt"
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(commit_message)
	if err2 != nil {
		log.Fatal(err2)
	}
	return err
}
