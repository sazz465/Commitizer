package main

import (
	"context"
	"flag"
	"fmt"

	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/iraj465/commitizer/helpers"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/rpcc"
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
	timeout     = flag.Int("timeout", 50, "Sets the context timeout value")
	pathCommits = flag.String("pathCommits", baseDir, "Path to store the commit files")
	pathCSV     = flag.String("pathCSV", baseDir, "Path to store the CSV file")
)

func main() {
	flag.Parse()

	relfPath, err := getRelativePath(*pathCommits)
	if err != nil {
		log.Fatal(err)
	}
	numAuthorCreated := make(map[string]int)  // map that stores number of commits created by each author
	numAuthorReviewed := make(map[string]int) // map that stores number of commits reviewed by each author

	// Function that uses helper funcs in helpers/ and does all the work
	err = commitizer_main(time.Duration(*timeout*int(time.Second)), relfPath, numAuthorCreated)
	if err != nil {
		log.Fatal(err)
	}

	// Parses commit metadata and creates contributions.csv
	err = helpers.Parser(relfPath, *pathCSV, numAuthorCreated, numAuthorReviewed)
	if err != nil {
		log.Fatal(err)
	}

}

// Gets relative path of the path passed in pathCommits flag with baseDir
func getRelativePath(pathCommits string) (string, error) {
	dirinfo, err := os.Stat(baseDir)
	if err != nil || !dirinfo.IsDir() {
		log.Println(err)
		os.MkdirAll(baseDir, 0755)
		fmt.Printf("\nDirectory %s is created\n", baseDir)

	}

	// Finds relative file path of the provided path flag with baseDir
	relfpath, err := filepath.Rel(baseDir, pathCommits)
	if err != nil {
		return relfpath, err
	}
	if relfpath == "." {
		relfpath = baseDir
	}
	return relfpath, nil
}

func commitizer_main(timeout time.Duration, relativeFilePath string, numAuthorCreated map[string]int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New("http://127.0.0.1:9222")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return err
		}
	}

	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return err
	}
	defer conn.Close() // Leaving connections open will leak memory.

	c := cdp.NewClient(conn)

	domLoadTimeout := 5 * time.Second
	err = helpers.Navigate(ctx, c.Page, *repoURL, domLoadTimeout)
	if err != nil {
		return err
	}

	var info DocumentInfo

	url, err := helpers.GetBranchURL(ctx, c, *branchName)
	if err != nil {
		return err
	}
	info.BranchURL = url

	// Navigate to main branch
	err = helpers.Navigate(ctx, c.Page, info.BranchURL, domLoadTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("\nNavigated to: %s\n", info.BranchURL)

	// Loop for getting *numCommits number of commits by calling getCommitAndMakeFile and making commit file every time
	commitIndex := 0
	for commitIndex < *numCommits {
		err = getCommitAndMakeFile(ctx, c, timeout, info, commitIndex, numAuthorCreated, relativeFilePath, domLoadTimeout)
		if err != nil {
			return err
		}

		commitIndex += 1
		fmt.Printf("Commit %d\n", commitIndex)
	}
	return nil
}

func getCommitAndMakeFile(ctx context.Context, c *cdp.Client, timeout time.Duration, info DocumentInfo, commitIndex int, numAuthorCreated map[string]int, relativeFilePath string, domLoadTimeout time.Duration) error {
	// Parse information from the document by evaluating JavaScript from expressions below.
	expression_commit_msg := `
	new Promise((resolve, reject) => {
		setTimeout(() => {
			const message = document.querySelector("body > div > div > pre").innerText;
			resolve({message});
		}, 500);
	});`

	expression_metadata := `
	new Promise((resolve, reject) => {
		setTimeout(() => {
			const commitHash = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(1) > td:nth-child(2)").innerHTML;
			const author = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(2) > td:nth-child(2)").innerText;
			const nextCommitHref = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(5) > td:nth-child(2) > a").href;
			const metadata = [commitHash,author,nextCommitHref]
			resolve({metadata});
		}, 500);
	});`

	commitMessage, details, err := helpers.CommitIterator(ctx, timeout, c, expression_commit_msg, expression_metadata, numAuthorCreated)
	if err != nil {
		return err
	}

	err = helpers.MakeCommitFile(commitMessage, details.Hash, relativeFilePath, commitIndex)
	if err != nil {
		return err
	}

	err = helpers.Navigate(ctx, c.Page, details.NextCommitHref, domLoadTimeout)
	if err != nil {
		return err
	}

	fmt.Printf("\nNavigated to: %s\n", details.NextCommitHref)
	return nil
}
