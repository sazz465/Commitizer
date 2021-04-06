# Commitizer

Commitizer automates generating commits and contributions table from [TAST](https://chromium.googlesource.com/chromiumos/platform/tast/) repository, with specified branch and other flags. See [Command line options](#command-line-options)

## Getting started

### Cloning go project

Assuming you have go set up in your system, please proceed as follows:

```
mkdir -p $GOPATH/src/github.com/iraj465/commitizer
cd $GOPATH/src/github.com/iraj465/commitizer

git clone git@github.com:iraj465/Commitizer.git
```
### Run headless chrome

```bash
docker container run -d -p 9222:9222 zenika/alpine-chrome --no-sandbox --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222

```
For headless chrome take a look [here](https://developers.google.com/web/updates/2017/04/headless-chrome)

To run the commitizer, either use already built executable `commitizer` in root and call it with specified [command line options](#command-line-options) if necessary.

For example, running 
```
./commitizer -repoURL=https://chromium.googlesource.com/chromiumos/platform/tast-tests/ -branchName=main -numCommits=5
```
will get the last 5 commits of the [repo](https://chromium.googlesource.com/chromiumos/platform/tast-tests/) from the branch `main`

If you wish to generate your own commitizer executable with name `exec-name`, run :

```
go build -o exec-name
```
The proceed with executable as outline above.

### Using Docker
For headless chrome docker image
```bash
docker container run -d --rm -p 9222:9222 zenika/alpine-chrome --no-sandbox --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222
```

For building commitizer_app
```bash
docker build -t commitizer_app .

docker run -it --rm --net host commitizer_app
```
---
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
