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
