# go‑api

![CodeRabbit Pull Request Reviews](https://img.shields.io/coderabbit/prs/github/dat1010/go-api?utm_source=oss&utm_medium=github&utm_campaign=dat1010%2Fgo-api&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews)
![Test Coverage](https://img.shields.io/codecov/c/github/dat1010/go-api)
![Go Version](https://img.shields.io/github/go-mod/go-version/dat1010/go-api)
![License](https://img.shields.io/github/license/dat1010/go-api)

This will be a simple API written in Golang. Mostly just trying to learn golang a little bit and do something fun.

Not sure what the direction of this api will be. starting simple with a /api/healthcheck at frist and growing from there.

The plan is to also use sqlite and deploy to AWS ECS.


Lets get started.

we are going to use GIN as our http client package?

## Prerequisites

Make sure your Go bin is on `$PATH`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Installing the `swag` CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

After that you can run from your project root:

```bash
swag init -g cmd/main.go
```

This will generate the `docs/` folder for your Swagger UI.

## Generate Docs

```
go run github.com/swaggo/swag/cmd/swag@latest init --generalInfo cmd/main.go --output docs
```

or run 

```
go generate ./cmd
```

## Testing

### Run Tests Locally

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Using the Test Script

We provide a convenient script that generates docs and runs tests:

```bash
./scripts/test.sh
```

This script will:
- Install dependencies
- Generate Swagger documentation
- Run all tests with coverage
- Display coverage summary

### Test Coverage

Test coverage is automatically calculated and uploaded to [Codecov](https://codecov.io) on every push and pull request. The current coverage is displayed in the badge above.

#### Setting up Codecov

1. Go to [Codecov.io](https://codecov.io) and sign in with your GitHub account
2. Add your repository to Codecov
3. Get your repository token from Codecov
4. Add the token as a GitHub secret named `CODECOV_TOKEN`
5. The badge will automatically update with your current coverage percentage

**Note**: The badge URL uses a placeholder token that Codecov will automatically replace with your actual token when you set up the integration.

### Running Specific Tests

```bash
# Run only controller tests
go test -v ./controllers

# Run tests matching a pattern
go test -v -run "TestGetPost" ./controllers

# Run tests with race detection
go test -race ./...
```
