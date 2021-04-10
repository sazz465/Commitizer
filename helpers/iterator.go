package helpers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/runtime"
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

// Evaluates javascript expressions;
// (1) `expressionCommitMessage` to get commit-message
// (2) `expressionMetadata` to get the commit-metadata
func CommitIterator(ctx context.Context, timeout time.Duration, c *cdp.Client, numAuthorCreated map[string]int) (string, CommitDetails, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var info DocumentInfo
	var details CommitDetails

	expressionCommitMessage := `
	new Promise((resolve, reject) => {
		setTimeout(() => {
			const message = document.querySelector("body > div > div > pre").innerText;
			resolve({message});
		}, 500);
	});`
	// Evaluates javascript expression `expressionCommitMessage`
	evalArgsCommitMessage := runtime.NewEvaluateArgs(expressionCommitMessage).SetAwaitPromise(true).SetReturnByValue(true)
	evalCommitMessage, err := c.Runtime.Evaluate(ctx, evalArgsCommitMessage)
	if err != nil {
		return info.CommitMessage, details, err
	}
	if err = json.Unmarshal(evalCommitMessage.Result.Value, &info); err != nil {
		return info.CommitMessage, details, err
	}

	expressionMetadata := `
	new Promise((resolve, reject) => {
		setTimeout(() => {
			const commitHash = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(1) > td:nth-child(2)").innerHTML;
			const author = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(2) > td:nth-child(2)").innerText;
			const nextCommitHref = document.querySelector("body > div > div > div.u-monospace.Metadata > table > tbody > tr:nth-child(5) > td:nth-child(2) > a").href;
			const metadata = [commitHash,author,nextCommitHref]
			resolve({metadata});
		}, 500);
	});`
	// Evaluates javascript expression `expressionMetadata`
	evalArgsMetadata := runtime.NewEvaluateArgs(expressionMetadata).SetAwaitPromise(true).SetReturnByValue(true)
	evalMetadata, err := c.Runtime.Evaluate(ctx, evalArgsMetadata)
	if err != nil {
		return info.CommitMessage, details, err
	}
	if err = json.Unmarshal(evalMetadata.Result.Value, &info); err != nil {
		return info.CommitMessage, details, err
	}

	// populates details: CommitDetails with data obtained from evaluating above javascript expressions
	details = CommitDetails{Hash: info.Metadata[0][:8], Author: info.Metadata[1], NextCommitHref: info.Metadata[2]}
	numAuthorCreated[details.Author]++

	return info.CommitMessage, details, nil
}
