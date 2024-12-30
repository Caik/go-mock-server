# Go Mock Server

Go Mock Server is a versatile tool crafted in Go to simplify the process of mocking HTTP requests, with a primary focus on being **user-friendly**, **powerful**, and **flexible**.

[![Build & Test](https://github.com/Caik/go-mock-server/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/Caik/go-mock-server/actions/workflows/build.yml)
[![Version](https://img.shields.io/github/release/Caik/go-mock-server.svg?style=flat-square)](https://github.com/Caik/go-mock-server/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/Caik/go-mock-server)](https://goreportcard.com/report/github.com/Caik/go-mock-server)
[![codecov](https://codecov.io/github/Caik/go-mock-server/graph/badge.svg)](https://codecov.io/github/Caik/go-mock-server)

Contents
========
- [Why Use Go Mock Server?](#-why-use-go-mock-server)
- [Key Features](#-key-features)
- [Installation](#-installation)
  - [Docker](#1-docker)
  - [Pre-compiled Binaries](#2-pre-compiled-binaries)
  - [Compiling Your Own Binary](#3-compiling-your-own-binary)
- [Usage](#-usage)
  - [Running the App](#1-running-the-app)
  - [Creating Mocks](#2-creating-mocks)
    - [a) Writing Mock Files](#a-writing-mock-files)
    - [b) Dynamic Mock Creation via API](#b-dynamic-mock-creation-via-api)
  - [Simulate Errors and Latencies](#3-simulate-errors-and-latencies)
  - [Integrate with Your Application](#4-integrate-with-your-application)
  - [Explore the Command-Line Options](#5-explore-the-command-line-options)
- [Want to Contribute?](#-want-to-contribute)
- [License](#%EF%B8%8F-license)

<br />

## 🤔 Why Use Go Mock Server?

Ever found yourself in situations where you needed to kick off development but the actual API wasn't ready? Or perhaps you faced challenges in confidently testing your application due to unreliable upstream services?

**Go Mock Server** steps in to tackle these common scenarios and more. Here's why it's your go-to solution:

### Rapid Development Kickstart

Need to mock an API response swiftly to jumpstart development when the real API isn't available? Go Mock Server lets you mock HTTP responses effortlessly without writing a single line of code.

### Robust Performance Testing

When conducting performance tests, confidence is key. Go Mock Server empowers you to simulate various network conditions, dynamically adjusting latency based on hosts and URIs. Ensure your application performs admirably under diverse response time scenarios.

### Reliable CI/CD Integration Testing

Integrate mock responses seamlessly into your CI/CD pipelines for thorough integration testing. Test your application's interactions with external APIs confidently, removing external interferences from your pipeline.

### Dynamic Configuration for Ultimate Flexibility

Whether you're simulating errors or latencies, Go Mock Server's Dynamic Configuration via API offers unparalleled flexibility. Fine-tune your mock server on-the-fly to adapt to evolving testing requirements.

If you've ever faced these challenges, or if you're just looking for a versatile and powerful HTTP mocking tool, you've come to the right place. Dive into the [usage](#-usage) section to discover how Go Mock Server can make your development and testing workflows smoother than ever.


<br />

## 📝 Key Features

**Go Mock Server** comes packed with a range of features designed to make HTTP request mocking easy, powerful, and flexible for your development and testing needs:

### 1. Easy Configuration

Configuring mock responses is a breeze. Simply write the desired response body in a file, and you're good to go. Optionally, utilize a dedicated API for dynamic configuration.

### 2. Latency Simulation

Simulate various network conditions by introducing latency to mock responses. Useful for testing the performance of your application under different scenarios.

### 3. Error Simulation

Mimic error responses to validate how your application handles unexpected situations. Ensure robustness and error-handling capabilities.

### 4. Host Resolution

The application automatically identifies mocks based on URI and HTTP method, streamlining host resolution for seamless integration with your applications. Define custom host configurations as needed.

### 5. Caching

Optimize performance and enhance Go Mock Server's reliability, making it more suitable for handling a high volume of requests, particularly beneficial in performance testing scenarios. 

### 6. Content-Type Awareness

Ensure accurate content-type handling by Go Mock Server. The application automatically returns the client's request content-type. In cases where no content-type is passed in the request, the application defaults to `text/plain`, ensuring seamless handling and compatibility with diverse APIs.

### 7. Dynamic Mock Creation

Dynamically create mocks on the fly in two convenient ways:

1. **File-Based Creation:** Simply create a new mock file or update/delete an existing mock file. Go Mock Server will automatically detect and apply these changes.

2. **API Interaction:** Interact with Go Mock Server's API to dynamically create, update, or delete mocks. This provides users with fine-grained control over mock configurations during runtime.

### 8. Dynamic Configuration via API

Configure the simulation of errors and latencies dynamically with Go Mock Server's powerful API. This feature empowers users to fine-tune error simulation and adjust latencies for specific hosts and/or URIs during runtime. Gain precise control over the testing environment to ensure comprehensive and targeted evaluations of your application's resilience and performance.

### 9. Cross-Platform Support

Go Mock Server provides precompiled binaries for Linux, Mac (AMD64 and ARM64), and Windows. Choose the binary that suits your platform or build from source if preferred.

Explore these features and more to streamline your API mocking workflow and accelerate your development process.

### 

<br />

## 💿 Installation

### 1. Docker

The easiest and recommended way to run **Go Mock Server** is via **Docker**: 

```bash
docker run --name mock-server --rm -p 8080:8080 -v $(pwd)/sample-mocks:/mocks caik/go-mock-server:latest --mocks-directory /mocks
```

Where `$(pwd)/sample-mocks` is the path in your host machine where you have stored the mocks files. In case you want to start the application without any pre-existing mock files, you can also omit it:

```bash
docker run --name mock-server --rm -p 8080:8080 caik/go-mock-server:latest --mocks-directory /mocks
```

Please also note that you change the port mapping from `8080` to any other port of your preference. 

### 2. Pre-compiled Binaries

Alternatively, you can download and run the already pre-compiled binaries. There are versions for **Linux**, **Mac**, and **Windows** on the **[Releases](https://github.com/Caik/go-mock-server/releases)** page.
So you only have to choose the appropriate file, download, extract it and run the binary in your machine.

PS: You may need to give execution permission to the binary after downloading it:

 ```shell
# giving execution permission on linux
chmod +x ./mock-server
```

### 3. Compiling Your Own Binary

If you have **Go** configured on your environment, you can also choose to build your own binary as well:

```shell
# building a MacOS on AMD64 binary
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o ./mock-server-darwin-amd64 cmd/mock-server/main.go
```

<br />

## 📖 Usage

### 1. Running the App

Please check [Installation](#-installation) for how to get/build the application binary. 

After you have the binary, you can run the server: 

```bash
# Example for Mac
./mock-server_mac --mocks-directory ./path-for-the-mocks-directory
```

### 2. Creating Mocks

To mock HTTP responses, you have two options:

#### a) Writing Mock Files:

Create mock files by writing the desired response body in a file within the specified mocks directory. Follow the naming convention for systematic organization: `{path-to-mocks-directory}/{host}/{uri-and-querystring}.{http-method}`

For example: `./path-for-the-mocks-directory/example.com/api/v1/resource.get`

Here's a breakdown of the components in the file name:

- `{path-to-mocks-directory}`: The directory where mock files are stored.
- `{host}`: The host name for which the mock is intended.
- `{uri-and-querystring}`: The URI and optional query string of the API endpoint.
- `{http-method}`: The HTTP method for which the mock is intended.

This convention allows for easy identification and management of specific mocks.

```bash
# Example for creating a mock file:
# GET example.com/api/v1/resource
echo '{"key": "value"}' > ./path-for-the-mocks-directory/example.com/api/v1/resource.get
```

#### b) Dynamic Mock Creation via API:

Alternatively, you can use the dedicated API to dynamically create mocks during runtime:

```bash
# Example for creating a mock via API:
# GET example.com/api/v1/resource
curl -X POST \
  -H "x-mock-host: example.host.com" \
  -H "x-mock-uri: /api/v1/resource" \
  -H "x-mock-method: GET" \
  --data-raw '{
    "key1": "value1",
    "key2": "value2"
  }' \
  http://localhost:8080/admin/mocks

```

To delete a mock:

```bash
# Example for deleting a mock via API:
# GET example.com/api/v1/resource
curl -X DELETE \
  -H "x-mock-host: example.host.com" \
  -H "x-mock-uri: /api/v1/resource" \
  -H "x-mock-method: GET" \
  http://localhost:8080/admin/mocks
```

For more details and additional API endpoints, please refer to the [Swagger documentation](https://github.com/Caik/go-mock-server/blob/main/docs/swagger.json).

### 3. Simulate Errors and Latencies

To enhance your testing experience, Go Mock Server provides powerful API endpoints for dynamically simulating errors and adjusting latencies. These features are particularly useful for testing your application's resilience under different conditions.

#### Simulate Errors
To simulate errors for a specific host, you can use the following example:

```bash
# Simulate a 500 Error for 20% of the requests to the host example.host.com
curl -X POST -H "Content-Type: application/json" -d '{
  "host": "example.host.com",
  "errors": {
    "500": {
      "percentage": 20
    }
  }
}' http://localhost:8080/admin/config/hosts/example.host.com/errors
```

#### Simulate Latency

```bash
# Simulating latency for the host example.host.com
curl -X POST -H "Content-Type: application/json" -d '{
  "host": "example.host.com",
  "latency": {
    "min": 100,
    "p95": 1800,
    "p99": 1900,
    "max": 2000
  }
}' http://localhost:8080/admin/config/hosts/example.host.com/latencies
```

For more details and additional API endpoints, please refer to the [Swagger documentation](https://github.com/Caik/go-mock-server/blob/main/docs/swagger.json).

### 4. Integrate with Your Application

Integrating Go Mock Server with your application is a straightforward process. By updating your application's URLs to point to Go Mock Server, you enable seamless testing and development. Go Mock Server intelligently identifies and sets the correct request host based on URI and HTTP method.

#### Example Scenario:

Let's consider a scenario where you have a web application that communicates with an external API. Initially, your application is configured to interact with the production API as follows:

```plaintext
Production API Base URL: https://example.host.com
```

Now, you want to test your application with different mock responses provided by Go Mock Server. Here's how you can integrate Go Mock Server into your testing environment:

##### 1. Update Application Configuration:

Update your application's configuration to point to the Go Mock Server URL:

```plaintext
Go Mock Server URL: http://localhost:8080
```

##### 2. Make Requests:

Your application can now make requests to Go Mock Server, and Go Mock Server will dynamically provide the mock responses based on the configured mocks.

For example, if your application originally made a request to:

```plaintext
GET https://example.host.com/data
```

Now the request will be:

```plaintext
GET http://localhost:8080/data
```

Go Mock Server will handle the request and respond according to the configured mocks.


### 5. Explore the Command-Line Options

To customize Go Mock Server's behavior, you can use the following command-line options:

| Option                  | Description                                                |
|-------------------------|------------------------------------------------------------|
| --mocks-directory       | Specify the directory for mock files.                      |
| --port                  | Set the port for the mock server. Default is 8080.         |
| --mocks-config-file     | Specify the path to a config file for additional settings. |
| --disable-cache         | Disable caching of responses.                              |
| --disable-latency       | Disable simulation of latency in responses.                |
| --disable-error         | Disable simulation of error responses.                     |

**Example:**

```bash
# Run the server on port 9090 with a custom mock directory and disable cache
./mock-server_mac --mocks-directory ./custom-mocks --port 9090 --disable-cache
```

<br />

## 🔧 Want to Contribute?

We welcome contributions from the community! If you're interested in helping improve Go Mock Server, please take a moment to review our [contribution guidelines](https://github.com/Caik/go-mock-server/blob/main/CONTRIBUTING.md).

<br />

## ⚖️ License

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/Caik/go-mock-server/blob/main/LICENSE)

Released 2023 by [Carlos Henrique Severino](https://github.com/Caik)
