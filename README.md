# Commitizer

Commitizer automates generating commits and contributions table from [TAST](https://chromium.googlesource.com/chromiumos/platform/tast/) repository, with specified branch and other flags. See [Command line options](#command-line-options)

---
## Getting started

### Cloning go project

Assuming you have go set up in your system, please proceed as follows:

```
mkdir -p $GOPATH/src/github.com/iraj465/commitizer
cd $GOPATH/src/github.com/iraj465/commitizer

git clone git@github.com:iraj465/Commitizer.git
```

### Using Docker


## Command line options


```
Usage of ./commitizer:

  -branchName string
        Name of the branch name on the first page to start the commitzer process (default "main")
  -numCommits int
        Number of commits to be obtained (default 10)
  -pathCSV string
        Path to store the CSV file (default "sample/commits-data/")
  -pathCommits string
        Path to store the commit files (default "sample/commits-data/")
  -repoURL string
        Repository URL to obtain the commits from (default "https://chromium.googlesource.com/chromiumos/platform/tast-tests/")
  -timeout int
        Sets the context timeout value (default 30)
```
