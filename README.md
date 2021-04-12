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
The proceed with executable as outlined above.

## Using Docker
For building commitizer with Headless chrome, run:
```bash
docker build -t commitizer .

docker run -it --rm --net host commitizer
```
---
## Command line options

Flag Name | Default value | Description |
---- | --- | --- |
-repoURL | https://chromium.googlesource.com/chromiumos/platform/tast-tests/ |Repository URL to obtain the commits from |
-numCommits | 10 | Number of commits to be obtained|
-branchName  | main | Name of the branch on the first page to start the commitizer process |
-timeout  | 15 | Sets the context timeout value|
-pathCommits |sample/commits-data/ | Path to store the commit files |
-pathCSV |sample/commits-data/ | Path to store the CSV file |
