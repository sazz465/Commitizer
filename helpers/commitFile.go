package helpers

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
)

// Makes commit file of the form `commit${commitIndex}_${commitHash}.txt`
func MakeCommitFile(commitMessage string, commitHash string, commitPath string, commitIndex int) error {

	path := commitPath + "/" + "commit" + strconv.Itoa(commitIndex+1) + "_" + commitHash + ".txt"
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "cannot create commit file")
	}
	defer f.Close()
	_, err2 := f.WriteString(commitMessage)
	if err2 != nil {
		return errors.Wrap(err, "cannot write commit message to file")
	}
	return err
}
