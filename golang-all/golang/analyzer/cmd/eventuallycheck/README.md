# `eventuallycheck`

`eventuallycheck` is a static analysis tool to detect [`gomega.Eventually`](https://godoc.org/github.com/onsi/gomega#Eventually) without [`Should`](https://godoc.org/github.com/onsi/gomega#Should) or [`ShouldNot`](https://godoc.org/github.com/onsi/gomega#ShouldNot).

## Usage

```console
$ eventuallycheck [FILES]
```

## Target functions

- [`Consistently`](https://godoc.org/github.com/onsi/gomega#Consistently)
- [`ConsistentlyWithOffset`](https://godoc.org/github.com/onsi/gomega#ConsistentlyWithOffset)
- [`Eventually`](https://godoc.org/github.com/onsi/gomega#Eventually)
- [`EventuallyWithOffset`](https://godoc.org/github.com/onsi/gomega#EventuallyWithOffset)
