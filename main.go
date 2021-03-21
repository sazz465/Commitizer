package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/runtime"
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
	myURL            = flag.String("repoURL", "https://chromium.googlesource.com/chromiumos/platform/tast-tests/", "Repository URL to obtain the commits from")
	numCommits       = flag.Int("numCommits", 10, "Number of commits to be obtained")
	branchName       = flag.String("branchName", "main", "Name of the branch name on the first page to start the commitzer process")
	timeout          = flag.Int("timeout", 30, "Sets the context timeout value")
	path             = flag.String("path", "commits/", "Path to store the commit files")
	baseDir          = "commits/" // default base directory for string commit files
	numAuthorCreated = make(map[string]int)
)

func main() {
	flag.Parse()

	dirinfo, err := os.Stat(baseDir)
	if err != nil || !dirinfo.IsDir() {
		log.Println(err)
		os.MkdirAll(baseDir, 0755)
		fmt.Printf("\nDirectory %s is created\n", baseDir)

	}
	relfpath, err := filepath.Rel(baseDir, *path)
	if err != nil {
		log.Fatal(err)
	}
	if relfpath == "." {
		relfpath = baseDir
	}

	err = run(time.Duration(*timeout*int(time.Second)), relfpath, numAuthorCreated)
	if err != nil {
		log.Fatal(err)
	}

	parser(relfpath)
	for author := range numAuthorReviewed {
		fmt.Printf("\nAuthor %s reviwed %d commits and created %d  \n", author, numAuthorReviewed[author], numAuthorCreated[author])
	}

}

func run(timeout time.Duration, relativeFilePath string, numAuthorCreated map[string]int) error {
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
	err = navigate(ctx, c.Page, *myURL, domLoadTimeout)
	if err != nil {
		return err
	}
	// fmt.Printf("Navigated to: %s\n", *myURL)

	/*
		Get branch URL of the `branchName` passed ar Command-line argument and navigate to it.
	*/
	// Parse information from the document by evaluating JavaScript.

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
	var info DocumentInfo

	url, err := getBranchURL(ctx, c, *branchName)
	if err != nil {
		return err
	}
	info.BranchURL = url

	// Navigate to main branch
	err = navigate(ctx, c.Page, info.BranchURL, domLoadTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("\nNavigated to: %s\n", info.BranchURL)

	commitIndex := 0
	var wg sync.WaitGroup
	wg.Add(*numCommits)
	for commitIndex < *numCommits {
		commitMessage, details, err := commit_iterator(ctx, timeout, c, expression_commit_msg, expression_metadata, numAuthorCreated)
		select {
		case <-ctx.Done():
			log.Println("\nInfo:Please consider increasing timeout")
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		go func(commitIndex int) {
			err = make_commit_file(commitMessage, details.Hash, relativeFilePath, commitIndex)
			if err != nil {
				log.Fatal(err)
			}
		}(commitIndex)

		err = navigate(ctx, c.Page, details.NextCommitHref, domLoadTimeout)
		if err != nil {
			return err
		}
		fmt.Printf("\nNavigated to: %s\n", details.NextCommitHref)

		commitIndex += 1
		fmt.Printf("Commit %d\n", commitIndex)

	}
	// for author := range numAuthorCreated {
	// 	fmt.Printf("\nAuthor %s created %d commits \n", author, numAuthorCreated[author])
	// }
	return nil
}

func getBranchURL(ctx context.Context, c *cdp.Client, requiredBranchName string) (string, error) {
	var info DocumentInfo
	childNodeIndex := 1

	branchNotFound := true
	for branchNotFound {
		expression_branch_url := fmt.Sprintf(`new Promise((resolve, reject) => {
			setTimeout(() => {
				const branchName = document.querySelector("body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(%d)").innerText;
				const branchURL = document.querySelector("body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(%d) > a").href;
				resolve({branchName,branchURL});
			}, 500);
		});`, childNodeIndex, childNodeIndex)

		evalArgs_branch_URL := runtime.NewEvaluateArgs(expression_branch_url).SetAwaitPromise(true).SetReturnByValue(true)
		eval_branch_URL, err := c.Runtime.Evaluate(ctx, evalArgs_branch_URL)
		if err != nil {
			return info.BranchURL, err
		}

		if err = json.Unmarshal(eval_branch_URL.Result.Value, &info); err != nil {
			return info.BranchURL, err
		}

		if info.BranchName == requiredBranchName {
			branchNotFound = false
		}
		childNodeIndex += 1
	}

	fmt.Printf("\nNavigated to branch branch with NAME : %q\n", info.BranchName)

	return info.BranchURL, nil

}

// navigate to the URL and wait for DOMContentEventFired. An error is
// returned if timeout happens before DOMContentEventFired.
func navigate(ctx context.Context, pageClient cdp.Page, url string, timeout time.Duration) error {
	// fmt.Printf("\nInside navigate")
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// Make sure Page events are enabled.
	err := pageClient.Enable(ctx)
	if err != nil {
		return err
	}

	// Open client for DOMContentEventFired to block until DOM has fully loaded.
	domContentEventFired, err := pageClient.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContentEventFired.Close()

	_, err = pageClient.Navigate(ctx, page.NewNavigateArgs(url))
	if err != nil {
		return err
	}

	_, err = domContentEventFired.Recv()
	return err
}
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

func commit_iterator(ctx context.Context, timeout time.Duration, c *cdp.Client, expression_commit_msg string, expression_metadata string, numAuthorCreated map[string]int) (string, CommitDetails, error) {
	// fmt.Printf("\ncommit iterator")
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var info DocumentInfo
	var details CommitDetails
	// Document commit msg
	evalArgs_commit_msg := runtime.NewEvaluateArgs(expression_commit_msg).SetAwaitPromise(true).SetReturnByValue(true)
	eval_commit_msg, err := c.Runtime.Evaluate(ctx, evalArgs_commit_msg)
	if err != nil {
		return info.CommitMessage, details, err
	}
	if err = json.Unmarshal(eval_commit_msg.Result.Value, &info); err != nil {
		return info.CommitMessage, details, err
	}
	// fmt.Printf("\nDocument commit msg is : %q\n", info.CommitMessage)

	// Get metadata table
	evalArgs_metadata := runtime.NewEvaluateArgs(expression_metadata).SetAwaitPromise(true).SetReturnByValue(true)
	eval_metadata, err := c.Runtime.Evaluate(ctx, evalArgs_metadata)
	if err != nil {
		return info.CommitMessage, details, err
	}
	if err = json.Unmarshal(eval_metadata.Result.Value, &info); err != nil {
		return info.CommitMessage, details, err
	}
	details = CommitDetails{Hash: info.Metadata[0][:8], Author: info.Metadata[1], NextCommitHref: info.Metadata[2]}
	numAuthorCreated[details.Author]++

	return info.CommitMessage, details, nil
}
