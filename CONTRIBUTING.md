# Contributing to Go Mock Server

Thank you for considering contributing to Go Mock Server! This guide outlines the steps to help you get started.

## How to Contribute

1. **Fork the Repository:** Start by forking the Go Mock Server repository to your GitHub account.

2. **Clone the Repository:** Clone the forked repository to your local machine using the following command:

```bash
git clone https://github.com/YourUsername/go-mock-server.git
```

3. **Create a Branch:** Create a new branch for your contribution:

```bash
git checkout -b feature/your-feature-name
```

4. **Make Changes:** Implement your changes or fix bugs. Ensure that your code follows the project's coding standards.

5. **Test Locally:** Test your changes locally to ensure they work as expected.

6. **Commit Changes:** Commit your changes with a clear and descriptive commit message:

```bash
git commit -m "Add feature: your feature name"
```

7. **Push Changes:** Push your changes to your GitHub repository:

```bash
git push origin feature/your-feature-name
```

8. **Open a Pull Request:** Submit a pull request with details about your changes. Our team will review your contribution.


<br />

## Code Style Guidelines

Please adhere to the project's coding standards when making changes. If you're unsure, feel free to ask for guidance.

<br />

## Reporting Issues

If you encounter any issues or have suggestions, please [open an issue](https://github.com/Caik/go-mock-server/issues) on the GitHub repository.

<br />

## Running Locally

### Prerequisites

Make sure you have the following installed:

- Go (at least version 1.21)
- Docker
- Docker Compose
- Make

### Instructions

1. Clone the repository:

```bash
git clone https://github.com/YourUsername/go-mock-server.git
cd go-mock-server
```

2. Build the project

```bash
go build ./...
```

3. Run the application:

```bash
go run cmd/mock-server/main.go --mocks-directory $(pwd)/sample-mocks 
```

<br />

## Update Documentation

Contributors are encouraged to keep the documentation up-to-date as they make changes or introduce new features. This ensures that the project remains well-documented and easy to understand for both existing and new users.

### Documentation Locations

- **Inline Comments:** Consider adding or updating comments within the code to provide context for your changes. This is especially helpful for complex logic or areas that might be confusing to others.

- **Separate Documentation Files:** If your changes introduce new features or modifications that require documentation, create or update separate documentation files. Follow any existing documentation structure and conventions.

### Contribution Checklist

Before submitting your pull request, ensure that you have considered the following regarding documentation:

1. **Explanatory Comments:** Are there sufficient comments in the code to explain the purpose and functionality of the changes?

2. **New Features or Changes:** Have you added or updated documentation for any new features or changes introduced in your contribution?

3. **README Updates:** If your changes impact user-facing features or configurations, make sure to update the relevant sections in the README file.

<br />

## Start Contributing Now!

Your contributions make Go Mock Server better for everyone. We appreciate your efforts and look forward to your valuable contributions!
