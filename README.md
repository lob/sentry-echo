# sentry-echo

`sentry-echo` provides a thin wrapper around the [raven-go](https://github.com/getsentry/raven-go) Sentry client, along with an Echo [error handler](https://echo.labstack.com/guide/error-handling) which reports internal server errors to Sentry.

## Installation

```
$ go get github.com/lob/sentry-echo/...
```

## Usage

To create a new Sentry client you must provide the DSN along with an `environment`; the environment will be included as a tag with all reports.

```go
import github.com/lob/sentry-echo/pkg/client

...

sentryClient, err := client.New(dsn)
```

If the client is unable to be created, an error will be returned.

After you've created a client you can register it as the error handler for your Echo application.

```go
import (
    github.com/lob/sentry-echo/pkg/client
    github.com/lob/sentry-echo/pkg/sentry
)

var e echo.Echo

...

sentryClient, _ := client.New(dsn)
sentry.RegisterErrorHandler(e, sentryClient)
```

## Reporting Errors to Sentry

The Sentry client may also be used to report events to Sentry other than internal server errors.

```go
sentryClient.Report(someError, nil)
```

If an `*http.Request` is available, it may be passed as the second parameter. If passed, Raven will extract the HTTP context from the request and include it with the Sentry report.

## Payload Sanitizing

The error handler and Sentry client will sanitize fields in request payloads, replacing them with `[CENSORED]`. To specify the fields to be censored, specify them when creating the Sentry client. 

For example, to filter `signature` and `credit_card`, create the client as follows:

```go,
client, err := client.NewWithOptions(
    client.Options{
        DSN: dsn,
        FilteredFields: []string{
            "signature",
            "credit_card",
        },
    },
)
```
