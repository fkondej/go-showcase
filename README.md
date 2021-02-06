# Go example

This repo contains a simple implementation of RESTful API (using `gin` library) and API client library in Go. The main focus is put on:
- BDD style tests (using `Ginkgo` and `GΩmega` libraries),
- `goroutines`, `channels` usage,
- Postgres database (using `pgx` library to connect),
- `docker-compose` setup,
- `VSCode` remote development setup,
- CI setup (using GitHub Actions).

## BDD tests

Tests are written in BDD style with [Ginkgo](https://onsi.github.io/ginkgo/) and [GΩmega](https://onsi.github.io/gomega/) libraries. Worth noting is usage of: [DescribeTable](https://onsi.github.io/ginkgo/#table-driven-tests) and [ghttp](https://onsi.github.io/gomega/#ghttp-testing-http-clients).

There are three types of tests:
- unit tests: files that end with `_unit_test.go`,
- integration tests using `docker-compose` setup: files that end with `_docker_test.go`,
- integration tests using `ghttp`: other files ending with `_test.go` (other than `_unit` and `_docker`).

Good examples:
- simple: [operation_delete_test.go](pkg/apiclient/operation_delete_test.go)
- more complex (e.g. timeout): [operation_negative_test.go](pkg/apiclient/operation_negative_test.go)

You can run tests with:
```bash
docker-compose up --exit-code-from workspace
```

## `goroutines`, `channels` usage

There is one usage in `FetchAll` function in [client.go](pkg/apiclient/client.go).

## Postgres database

`apiserver` reads and stores its data in Postgres database. The `pgx` (`pgxpool`) library is used to communicate with it. Few interesting things:
- implemented `UPSERT` operation with `INSERT ... ON CONFLICT` [account_service.go](pkg/apiserver/account_service.go),
- implemented query a JASON column [account_service.go](pkg/apiserver/account_service.go),
- used a transaction to insert a large number of generated data [test_helper_db.go](pkg/libtest/test_helper_db.go)

## `docker-compose` setup

Three containers: Postgres `db`, `apiserver`, and `workspace`. [docker-compose.go](docker-compose.yml). `apiserver` container uses `CompileDaemon` to observer `apiserver` source code and recompile+rerun on server code change.

## VSCode setup

[.devcontainer](.devcontainer).

## CI GitHub Actions

## Other

#### apiclient usage

```go
// new client
config = orgaccount.AccountClientConfig{
    URL:      serverURL.String(),
    Timeout:  time.Second,
}
client = orgaccount.NewAccountClient(&config)

// Create operation
accountData, err := client.Create(
    "ad27e265-9605-4b4b-a0e5-3003ea9cc4dc",
    "eb0bd6f5-c3f5-44b2-b677-acd23cdde73c",
    &orgaccount.AccountAttributes{
        Country: "GB",
        // ...
    },
)

// Fetch operation
accountData, err := client.Fetch("ad27e265-9605-4b4b-a0e5-3003ea9cc4dc")

// List operation
accountPages := client.List(orgaccount.AccountPage{
    PageNumber: 3,
    PageSize: 50,
})

// option 1: Page by Page
for accountPages.Next() {
    accountDataList, err := accounts.Data()
}
// option 2: Account by Account
for acc := range accounts.FetchAll() {
    // ...
}

// Delete operation
deleted, err := accountClient.Delete("ad27e265-9605-4b4b-a0e5-3003ea9cc4dc", 3)
				
```
