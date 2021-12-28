# zenpull
A Go CLI tool for downloading all the files from a list of urls in parallel

## Features
This tool may be smol, but is packed with a ton of features:
- Download files as fast as it can
- Downloads them in parallel
- Organizes files based on file types
- Gives every file a unique name
- Stands robust against unstable internet connections
- Uses http ranged requests (whenever it can) to:
  - Make sure files are downloaded 100%
  - Avoid re-downloading parts of files
- Makes sure downloads are idempotent, which means:
  - Every url is only downloaded once even if you call zenpull on the same list twice
- Doesn't create any database or BS to remember what files were downloaded
- Requires no configurations at all
- The number of green threads, used to download files, are customizable

What it can not do, however, is:
- Get you a girlfriend
- or, make you a sandwich even if you use sudo

## Installation
Make sure you have `go` installed. If not, install it from [https://go.dev/dl](https://go.dev/dl). Then run the following command:
```
go install github.com/keogami/zenpull@latest
```
and now you have it :3

## Usage
Store the list of urls in a file (say `list.txt`) with a single url on each line:
```
https://example.com/file1.txt
https://exmaple.com/file2.png
https://someotherexample.com/path/to/file.svg
```

Then, just pass zenpull the list of urls that you wanna download, like so:
```
zenpull list.txt
```
and you should have your files downloaded in the current directory.

By default, zenpull downloads 5 files in parallel. However, you can fine tune that number using the `-worker` flag:
```
zenpull -worker=10 list.txt
```
----

# Cheers?
Buy me a beer later~ ;3
