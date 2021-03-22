package helpers

import (
	"log"
	"os"
	"strconv"
)

func make_commit_file(commit_message string, commit_hash string, fpath string, commitIndex int) error {

	// fmt.Printf("\nstarted Commit file for commit %d", commitIndex+1)
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
	// fmt.Printf("\nFinished file for commit %d", commitIndex+1)
	return err
}
