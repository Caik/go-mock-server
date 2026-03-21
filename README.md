![go-mock-server](.github/banner.png)

# Go Mock Server

[![Build & Test](https://github.com/Caik/go-mock-server/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/Caik/go-mock-server/actions/workflows/build.yml)
[![Version](https://img.shields.io/github/release/Caik/go-mock-server.svg?style=flat-square)](https://github.com/Caik/go-mock-server/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/Caik/go-mock-server)](https://goreportcard.com/report/github.com/Caik/go-mock-server)
[![codecov](https://codecov.io/github/Caik/go-mock-server/graph/badge.svg)](https://codecov.io/github/Caik/go-mock-server)

**Go Mock Server** is a lightweight HTTP mock server built in Go. Run it locally, point your app at it, and control every response — no real API needed.

Ever found yourself waiting for a backend that isn't ready? Dealing with flaky third-party services in CI? Trying to reproduce a rate-limit or 503 error that only happens in production? Go Mock Server solves all of that: define your mock responses as plain files, start the server, and your app has a fully controllable API to talk to — complete with a web UI for managing everything in real time.

## Contents

- [Quick Start](#-quick-start)
- [How It Works](#-how-it-works)
- [Installation](#-installation)
  - [Docker](#1-docker)
  - [Pre-compiled Binaries](#2-pre-compiled-binaries)
  - [Compiling Your Own Binary](#3-compiling-your-own-binary)
- [Creating Mocks](#-creating-mocks)
  - [Mock Files](#a-mock-files)
  - [Dynamic Creation via API](#b-dynamic-creation-via-api)
- [Simulate Latency and Status Codes](#-simulate-latency-and-status-codes)
  - [Latency Simulation](#latency-simulation)
  - [Status Code Simulation](#status-code-simulation)
- [Integrate with Your Application](#-integrate-with-your-application)
- [Admin UI](#%EF%B8%8F-admin-ui)
- [Command-Line Options](#-command-line-options)
- [Want to Contribute?](#-want-to-contribute)
- [License](#%EF%B8%8F-license)
