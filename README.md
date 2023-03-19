# Go Mock Server

Go Mock Server is a tool built with Go to support HTTP requests mocking with a main goal in mind: To be very **easy to use** along with being very **powerful** and **flexible**. 

<br />

Contents
========
 * [Why?](#-why)
 * [Features](#-features)
 * [Installation](#-installation)
 * [Usage](#-usage)
 * [Want to contribute?](#-want-to-contribute)
 * [License](#%EF%B8%8F-license)

<br />

## 🤔 Why?

Have you ever needed to mock an API response to start the development because the actual API was not yet ready to use? 

What about that performance testing you needed to do in our app but you could not confidently rely upon your upstream services? 

Need to mock some external API in your CI/CD pipeline for integration testing purposes?

If you have seen yourself in any kind of situation like that, so you have come to the right place!! 

**Go Mock Server** aims to solve all of these issues by allowing you to quickly set up HTTP mocks without having to write a single line of code! 

The only thing you need to do is to write the desired response body in a file and you are ready to go! Alternatively, you also have the option to use a dedicated API to set up and configure your mocks! 

So please check out the [usage](#usage) section so you can fully understand how this neat tool can greatly help you easily to fulfill these tasks. 


<br />

## 📝 Features
WIP

<br />

## 💿 Installation

There is already a compiled binary for **[Linux](https://github.com/Caik/go-mock-server/blob/main/dist/mock-server_linux)**, **[Mac](https://github.com/Caik/go-mock-server/blob/main/dist/mock-server_linux)** and another one for **[Windows](https://github.com/Caik/go-mock-server/blob/main/dist/mock-server.exe)** on the **dist/** directory.
So you only have to download the appropriate binary and run on your machine.

PS: You may need to give execution permission to the binary after downloading it:

 ```shell
# giving execution permission on linux
chmod +x ./mock-server_linux
```

If you have **Go** configured on your environment, you can build your own binaries as well:

```shell
# building a MacOS on AMD64 binary
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o ./mock-server-darwin-amd64 cmd/mock-server/main.go
```

<br />

## 📖 Usage

Example:

```shell
# starting the server
./mock-server_mac --mocks-directory ./path-for-the-mocks-directory
```

<br />

## 🔧 Want to Contribute?

Please take a look at our [contributing](https://github.com/Caik/go-mock-server/blob/main/CONTRIBUTING.md) guidelines if you're interested in helping!

<br />

## ⚖️ License

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/Caik/go-mock-server/blob/main/LICENSE)

Released 2023 by [Carlos Henrique Severino](https://github.com/Caik)