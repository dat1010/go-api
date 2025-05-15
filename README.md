# go‑api

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
