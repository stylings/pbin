# pbin - privatebin cli tool
---

this is a fork of [cbluth/pbin](https://github.com/cbluth/pbin). the original pbin is unmaintained; its last commit was on 2021-05-10.

this is still a small tool for putting pastes on privatebin. it encrypts the paste locally, uploads ciphertext, and keeps the key in the url after `#`.

this fork updates it for the current privatebin json api and format v2 encryption. public instances are listed here: https://privatebin.info/directory/

if you want a maintained, feature-rich privatebin cli, use [gearnode/privatebin](https://github.com/gearnode/privatebin). this fork exists for the old `pbin` workflow, minimal deps, and a smaller binary.

# Install:

install the normal "go" way:

```
go install github.com/stylings/pbin/cmd/pbin@latest
```

or build from this checkout:

```
go build -o pbin ./cmd/pbin
```

# Basic Usage:

Upload Paste:

```
$ echo "anything" | pbin
https://privatebin.net/?5f9fc3956e8bc7bd#8NBafBFyqKWZrqPHiw4hC1JkL9Vx9mxEUGtXBT5wLNJF
```

Download Paste:

```
$ URL="https://privatebin.net/?908a9812a167d638#AKQaAp7bwC9t7gLBJkLXxJt1ZQQyW4bfjnBCzbn73c95"
$ pbin $URL
## prints content to stdout
```

## Advanced Usage

You can set additional options if some of these arguments, only when creating a paste:

- -burn
- -open
- -base64
- -password

Upload Base64 Paste:

```
$ cat cat-meme.gif | pbin -base64
```

Download Base64 Paste:

```
$ pbin $URL -base64 > cat-meme.gif
```

Download Base64 Paste to file:

```
$ pbin $URL -base64 -o cat-meme.gif
```

Upload Paste with Burn After Read Once:

```
$ echo "anything" | pbin -burn
```

Download Paste to filepath:

```
$ pbin $URL -output cat-meme.gif
```

Upload Paste with discussion forum enabled:

```
$ echo "anything" | pbin -open
```

Upload Paste with password protection:

```
$ echo "anything" | pbin -password mySecretPassw0rd
```

## Expiry Options

You can set the expiry with one of these arguments, only when creating a paste:

- -hour
- -day
- -week
- -month
- -year
- -never

examples:

```
$ echo "anything" | pbin -hour # <- will expire after 1 hour
$ echo "anything" | pbin -week # <- will expire after 1 week
$ echo "anything" | pbin -month # <- will expire after 1 month
$ cat cat-meme.gif | pbin -base64 -never
```
