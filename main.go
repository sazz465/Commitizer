package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	BranchURL     string   `json:"branch"`
	CommitMessage string   `json:"message"`
	Metadata      []string `json:"metadata"`
}

var (
	myURL      = "https://chromium.googlesource.com/chromiumos/platform/tast-tests/"
	numCommits = 10
)

func main() {
	args := os.Args
	fmt.Println(args)
	err := run(10 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
}

func run(timeout time.Duration) error {
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

	domLoadTimeout := 50 * time.Second
	err = navigate(ctx, c.Page, myURL, domLoadTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("Navigated to: %s\n", myURL)

	/*
		Get `main` branch URL and navigate to it. (Assuming stable JsPATH)
	*/
	// Parse information from the document by evaluating JavaScript.
	expression_branch_url := `
		new Promise((resolve, reject) => {
			setTimeout(() => {
				const branch = document.querySelector("body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(1) > a").href;
				resolve({branch});
			}, 500);
		});`

	expression_commit_msg := `
	new Promise((resolve, reject) => {
		setTimeout(() => {
			const message = document.querySelector("body > div > div > pre").innerHTML;
			resolve({message});
		}, 500);
	});`

	expression_metadata := `
	new Promise((resolve, reject) => {
		setTimeout(() => {
			const commitHash = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(1) > td:nth-child(2)").innerHTML;
			const author = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(2) > td:nth-child(2)").innerHTML;
			const nextCommitHref = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(5) > td:nth-child(2) > a").href;
			const metadata = [commitHash,author,nextCommitHref]
			resolve({metadata});
		}, 500);
	});`

	// Document Main branch URL
	evalArgs_branch_URL := runtime.NewEvaluateArgs(expression_branch_url).SetAwaitPromise(true).SetReturnByValue(true)
	eval_branch_URL, err := c.Runtime.Evaluate(ctx, evalArgs_branch_URL)
	if err != nil {
		return err
	}
	var info DocumentInfo
	if err = json.Unmarshal(eval_branch_URL.Result.Value, &info); err != nil {
		return err
	}
	// fmt.Printf("\nDocument Main branch URL : %q\n", info.BranchURL)

	// Navigate to main branch
	err = navigate(ctx, c.Page, info.BranchURL, domLoadTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("\nNavigated to: %s\n", info.BranchURL)

	commitIndex := 0
	for commitIndex < numCommits {
		commitMessage, details, err := commit_iterator(ctx, timeout, c, expression_commit_msg, expression_metadata)
		select {
		case <-ctx.Done():
			log.Println("\nInfo:Please consider increasing timeout")
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		err = make_commit_file(commitMessage, details.Hash)
		if err != nil {
			return err
		}

		// go make_commit_file(commitMessage, details.Hash)

		err = navigate(ctx, c.Page, details.NextCommitHref, domLoadTimeout)
		if err != nil {
			return err
		}
		fmt.Printf("\nNavigated to: %s\n", details.NextCommitHref)

		commitIndex += 1
		fmt.Println(commitIndex)

	}

	return nil
}

// navigate to the URL and wait for DOMContentEventFired. An error is
// returned if timeout happens before DOMContentEventFired.
func navigate(ctx context.Context, pageClient cdp.Page, url string, timeout time.Duration) error {
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
func make_commit_file(commit_message string, commit_hash string) error {
	f, err := os.Create(commit_hash + ".txt")
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

func commit_iterator(ctx context.Context, timeout time.Duration, c *cdp.Client, expression_commit_msg string, expression_metadata string) (string, CommitDetails, error) {
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
	// fmt.Printf("\nMetadata Table: %q\n", info.Metadata)

	details = CommitDetails{Hash: info.Metadata[0][:8], Author: info.Metadata[1], NextCommitHref: info.Metadata[2]}

	// fmt.Printf("\nMetadata Table: %q\n", details)

	return info.CommitMessage, details, nil
}

// func get_commit(ctx context.Context, )  {

// }
