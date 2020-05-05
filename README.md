# Twitter Filtered Stream API Client

A simple Go program to consume [Twitter's filtered stream API](https://developer.twitter.com/en/docs/labs/filtered-stream/overview).

## Settings

This program depends on two environment variables to work (you can copy `.env.sample` as `.env` as well):

* `TWITTER_API_KEY`
* `TWITTER_API_SECRET`

There are also two optional environment variables:

* `TWITTER_APP_NAME` which is the name of your Twitter application (used as _user-agent_ in the API requests)
* `TWITTER_ACCESS_TOKEN` which is **highly recommended** since Twitter discourage you to use the endpoint that grants this token (read further on to know how to get this token using this client)

## Usage

You can either run the commands with `go run main.go` or calling the binary generated with `go build main.go`.

### 1. Get your OAuth2 token

```console
$ go run main.go token
```

This command displays the token to be used for authentication. It's **highly recommended** to save this value as an environment variable called ``TWITTER_ACCESS_TOKEN`` (or save it in your `.env`).

### 2. Start to listen to the stream

```console
$ go run main.go stream
```

This command starts to listen to the stream of tweets. Start this program in an instance of the terminal and then move to another instance to create your rules – and wait to see the magic! Once a tweet matches the rules you create in the next step, it will be saved in the `data/` directory. 

### 3. Manage your rules

```console
$ go run main.go rule ls
```

This command lists the existing rules for the filtered stream.

```console
$ go run main.go rule new --query YOUR-SEARCH-TERM-HERE
```

This command creates a new rule using the text passed with the `--query` option.

```console
$ go run main.go rule rm
```

This command deletes **all** existing rules.

## Contributing

I'm still learning Go, so any feedback, issue, comment or pull request is welcome!

I know we have libraries to handle the requests to Twitter API, but this bare implementation was a way I've found to study `net/http`, ok?