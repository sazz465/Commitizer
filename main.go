package main

import (
	"context"
	"flag"
	"fmt"

	"log"
	"os"
	"time"

	"github.com/iraj465/commitizer/helpers"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/rpcc"
	"github.com/pkg/errors"
)

type CommitDetails struct {
	Hash           string `json:"hash"`
	Author         string `json:"author"`
	NextCommitHref string `json:"nexthash"`
}

type DocumentInfo struct {
	BranchURL     string   `json:"branchURL"`
	BranchName    string   `json:"branchName"`
	CommitMessage string   `json:"message"`
	Metadata      []string `json:"metadata"`
}

// Command line variables
var (
	baseDir     = "sample/commits-data/" // default base directory for string commit files
	repoURL     = flag.String("repoURL", "https://chromium.googlesource.com/chromiumos/platform/tast-tests/", "Repository URL to obtain the commits from")
	numCommits  = flag.Int("numCommits", 10, "Number of commits to be obtained")
	branchName  = flag.String("branchName", "main", "Name of the branch on the first page to start the commitizer process")
	timeout     = flag.Int("timeout", 15, "Sets the context timeout value")
	pathCommits = flag.String("pathCommits", baseDir, "Path to store the commit files")
	pathCSV     = flag.String("pathCSV", baseDir, "Path to store the CSV file")
)

func main() {
	flag.Parse()

	outputFlagValues()

	commitPath, err := getPath(*pathCommits)
	if err != nil {
		log.Fatal(err)
	}
	numAuthorCreated := make(map[string]int)  // map that stores number of commits created by each author
	numAuthorReviewed := make(map[string]int) // map that stores number of commits reviewed by each author

	err = commitizerMain(time.Duration(*timeout*int(time.Second)), commitPath, numAuthorCreated)
	if err != nil {
		log.Fatal(err)
	}

	csvPath, err := getPath(*pathCSV)
	if err != nil {
		log.Fatal(err)
	}
	err = helpers.Parser(commitPath, csvPath, numAuthorCreated, numAuthorReviewed)
	if err != nil {
		log.Fatal(err)
	}

}

// Lists all passed/default flag values
func outputFlagValues() {
	fmt.Println("---------------------------------------------")
	fmt.Printf("\nRepo URL\t\t: %s\n", *repoURL)
	fmt.Printf("Commits Num\t\t: %d\n", *numCommits)
	fmt.Printf("Repo Branch Name\t: %s\n", *branchName)
	fmt.Printf("Timeout \t\t: %d\n", *timeout)
	fmt.Printf("Path for Commmits\t: %s\n", *pathCommits)
	fmt.Printf("Path for CSV\t\t: %s\n", *pathCSV)

	fmt.Println("---------------------------------------------")
}

// Creates path directory if not already present
func getPath(path string) (string, error) {
	dirinfo, err := os.Stat(path)
	if err != nil || !dirinfo.IsDir() {
		log.Println(err)
		os.MkdirAll(path, 0755)
		fmt.Printf("\nDirectory %s is created\n", path)

	}
	return path, nil
}

// Function that uses helper funcs in helpers/ and does all the work
// of getting numCommits number of commits and makes corresponding commit files
func commitizerMain(timeout time.Duration, commitsPath string, numAuthorCreated map[string]int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New("http://127.0.0.1:9222")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return errors.Wrap(err, "couldn't create cdp target")
		}
	}

	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return errors.Wrap(err, "couldn't initiate new rpc connection with cdp target")
	}
	defer conn.Close() // Leaving connections open will leak memory.

	c := cdp.NewClient(conn)

	domLoadTimeout := 5 * time.Second
	err = helpers.Navigate(ctx, c.Page, *repoURL, domLoadTimeout)
	if err != nil {
		return err
	}

	var info DocumentInfo

	err = navigateToBranch(ctx, c, info, *branchName, domLoadTimeout)
	if err != nil {
		return err
	}

	// Loop for getting *numCommits number of commits by calling getCommitAndMakeFile and making commit file every time
	commitIndex := 0
	for commitIndex < *numCommits {
		err = getCommitAndMakeFile(ctx, c, info, commitIndex, numAuthorCreated, commitsPath, domLoadTimeout)
		if err != nil {
			return err
		}

		commitIndex += 1
		fmt.Printf("Commit %d obtained!\n", commitIndex)
	}
	return nil
}

// Navigates to branch with name `branchName` (default branchName is "main")
func navigateToBranch(ctx context.Context, c *cdp.Client, info DocumentInfo, branchName string, domLoadTimeout time.Duration) error {
	url, err := helpers.GetBranchURL(ctx, c, branchName)
	if err != nil {
		return err
	}
	info.BranchURL = url

	// Navigate to main branch
	err = helpers.Navigate(ctx, c.Page, info.BranchURL, domLoadTimeout)
	if err != nil {
		return err
	}
	return nil
}

// Gets the commit which is the commitIndex(^th) commit
// and makes a commit file (.txt), and then navigates to the next commit page
func getCommitAndMakeFile(ctx context.Context, c *cdp.Client, info DocumentInfo, commitIndex int, numAuthorCreated map[string]int, commitsPath string, domLoadTimeout time.Duration) error {

	commitMessage, details, err := helpers.CommitIterator(ctx, c, numAuthorCreated)
	if err != nil {
		return err
	}

	err = helpers.MakeCommitFile(commitMessage, details.Hash, commitsPath, commitIndex)
	if err != nil {
		return err
	}

	err = helpers.Navigate(ctx, c.Page, details.NextCommitHref, domLoadTimeout)
	if err != nil {
		return err
	}

	return nil
}
