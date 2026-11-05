# sentry-echo

`sentry-echo` provides a thin wrapper around the [raven-go](https://github.com/getsentry/raven-go) Sentry client, along with an Echo [error handler](https://echo.labstack.com/guide/error-handling) which reports internal server errors to Sentry.

## Installation

```
$ go get github.com/lob/sentry-echo/v2/...
```

v2 requires `github.com/labstack/echo/v5`; use v1 with echo v4.

## Usage

To create a new Sentry client you must provide the DSN along with an `environment`; the environment will be included as a tag with all reports.

```go
import "github.com/lob/sentry-echo/v2/pkg/sentry"

...

sentryClient, err := sentry.New(dsn)
```

If the client is unable to be created, an error will be returned.

After you've created a client you can register it as the error handler for your Echo application.

```go
import (
    "github.com/labstack/echo/v5"
    sentryecho "github.com/lob/sentry-echo/v2/pkg"
    "github.com/lob/sentry-echo/v2/pkg/sentry"
)

...

e := echo.New()
sentryClient, _ := sentry.New(dsn)
sentryecho.RegisterErrorHandler(e, &sentryClient)
```

Responses keep the status of any error that implements `echo.HTTPStatusCoder`, including echo's own errors such as `echo.ErrNotFound`. Only 500s are reported to Sentry.

## Reporting Errors to Sentry

The Sentry client may also be used to report events to Sentry other than internal server errors.

```go
sentryClient.Report(someError, nil)
```

If an `*http.Request` is available, it may be passed as the second parameter. If passed, Raven will extract the HTTP context from the request and include it with the Sentry report.

## Payload Sanitizing

The error handler and Sentry client will sanitize fields in request payloads, replacing them with `[CENSORED]`. Cookies often contain session tokens; as such the Cookies are **always** censored. The client can also censor query string fields and sensitive headers.

To specify query string fields to be censored, specify them when creating the Sentry client.

For example, to filter `signature` and `credit_card`, create the client as follows:

```go,
client, err := sentry.NewWithOptions(
    sentry.Options{
        DSN: dsn,
        FilteredFields: []string{
            "signature",
            "credit_card",
        },
    },
)
```

If the client is created via `sentry.New()`, the `Authorization` header will be censored. An alternative list of headers to censor may be provided as part of the `sentry.Options`.
