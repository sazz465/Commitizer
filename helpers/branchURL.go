package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/runtime"
)

// Get branch URL of the `branchName` passed as Command-line flag and navigates to it.
func GetBranchURL(ctx context.Context, c *cdp.Client, requiredBranchName string) (string, error) {
	var info DocumentInfo
	childNodeIndex := 1

	branchNotFound := true
	for branchNotFound {
		expressionBranchURL := fmt.Sprintf(`new Promise((resolve, reject) => {
			setTimeout(() => {
				const branchName = document.querySelector("body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(%d)").innerText;
				const branchURL = document.querySelector("body > div > div > div.RepoShortlog > div.RepoShortlog-refs > div > ul > li:nth-child(%d) > a").href;
				resolve({branchName,branchURL});
			}, 500);
		});`, childNodeIndex, childNodeIndex)

		evalArgsBranchURL := runtime.NewEvaluateArgs(expressionBranchURL).SetAwaitPromise(true).SetReturnByValue(true)
		evalBranchURL, err := c.Runtime.Evaluate(ctx, evalArgsBranchURL)
		if err != nil {
			return info.BranchURL, err
		}

		if err = json.Unmarshal(evalBranchURL.Result.Value, &info); err != nil {
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
