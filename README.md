# SafariBooks Downloader

SafariBooks downloader is a project used to download ebook from https://www.safaribooksonline.com/ and convert it to be epub.
This project mimics https://github.com/nicohaenggi/SafariBooks-Downloader by using golang.

[![Go Report Card](https://goreportcard.com/badge/github.com/kkc/safari-books-downloader)](https://goreportcard.com/report/github.com/kkc/safari-books-downloader)

# Installation

```
go get -d github.com/kkc/safari-books-downloader
cd ${GOPATH:-$HOME/go}/src/github.com/kkc/safari-books-downloader
go build -o safari-downloader
```

# Usage example

```
safari-downloader bookId

Usage:
safari-downloader bookId [flags]

Flags:
-h, --help              help for safari-downloader
-o, --output string     output path the epub file should be saved to (default "ebook.epub")
-p, --password string   password of the SafariBooksOnline user
-u, --username string   username of the SafariBooksOnline user - must have a **paid/trial membership**, otherwise will not be able to access the books
```


# Config

you could set up your own username and password in the local config file instead of entering username and password everytime.
put your username and password in the `~/.safari.toml`.

```
[safari]
username = ""
password = ""
```

# Development setup

# Release History

* 0.0.1
    * Work in progress

## Meta

Your Name – [@kakashiliu](https://twitter.com/kakashiliu)
Distributed under the Apache License 2.0. See ``LICENSE`` for more information.

[https://github.com/kkc](https://github.com/kkc/)
