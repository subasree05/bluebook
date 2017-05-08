# bluebook

API test management and execution.

## What is Bluebook?

Inspired by Terraform, Bluebook let's you reclare and run API tests for your
services.

## Installing

Currently you can only install bluebook from source code:

```
go get -u github.com/bluebookrun/bluebook
```

## Example

You can see some examples in `regressions/regressions.bcl`.

In short you can write API tests that look something like this:

```
# This is an exmaple showing how you can login and make an authenticated request

resource "http_assertion_status_code" "equals_200" {
    equals = "200"
}

resource "http_outlet_json_field" "api_key" {
    path = "data.api_key"
    variable = "api_key"
}

resource "http_step" "authenticate" {
    method = "POST"
    url = "http://localhost:12345/authenticate"
    body = <<<EOF
username=username&password=password
EOF

    assertions = [
        "${http_assertion_status_code.equals_200.id}",
    ]

    outlets = [
        "${http_outlet_json_field.api_key.id}",
    ]
}

resource "http_step" "get_document" {
    method = "GET"
    url = "http://localhost:12345/document/1?api_key=${var.api_key}"

    assertions = [
        "${http_assertion_status_code.equals_200.id}",
    ]
}

resource "http_test" "login_and_get_document" {
    steps = [
        "${http_step.authenticate.id}",
        "${http_step.get_document.id}",
    ]
}
```
