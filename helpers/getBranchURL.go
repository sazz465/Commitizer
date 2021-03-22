package helpers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/runtime"
)

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
