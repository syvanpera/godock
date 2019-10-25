# GoDock - A console flowdock client

## Installation

Go to https://www.flowdock.com/oauth/applications and create a new application
and grab the Application ID and Secret from there.

Create a file `config.toml` to the root of the project and add the following
lines:

```
[Flowdock]
ClientId = "<YOUR APP ID>"
ClientSecret = "<YOUR APP SECRET>"
```

Then just run:
`go run main.go`
It'll prompt you for the authorization for the first time you run it.

Use `go run main.go -d` if you want to see a bunch of debug info.
