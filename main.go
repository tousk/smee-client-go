package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/buger/jsonparser"
	flags "github.com/jessevdk/go-flags"
)

const (
	VERSION = "1.0"
)

// ValidMAC reports whether messageMAC is a valid HMAC tag for message.
func ValidMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func hex2bytes(hexstr string) []byte {
	src := []byte(hexstr)
	dst := make([]byte, hex.DecodedLen(len(src)))
	_, err := hex.Decode(dst, src)
	if err != nil {
		log.Fatal(err)
	}
	return dst
}

/*
Usage: smee [options]

Options:
  -v, --version          output the version number
  -u, --url <url>        URL of the webhook proxy service. Default: https://smee.io/new
  -t, --target <target>  Full URL (including protocol and path) of the target service the events will forwarded to. Default: http://127.0.0.1:PORT/PATH
  -p, --port <n>         Local HTTP server port (default: 3000)
  -P, --path <path>      URL path to post proxied requests to` (default: "/")
  -h, --help             output usage information
*/

type Options struct {
	Version []bool `short:"v" long:"version" description:"output the version number"`
	URL     string `short:"u" long:"url" description:"URL of the webhook proxy service. Required." required:"yes"`
	Target  string `short:"t" long:"target" description:"Full URL (including protocol and path) of the target service the events will forwarded to. Required." required:"yes"`
	Secret  string `short:"s" long:"secret" description:"Secret to be used for HMAC-SHA1 secure hash calculation"`
}

var opts Options
var parser = flags.NewParser(&opts, flags.HelpFlag)

func main() {
	_, err := parser.Parse()

	if len(opts.Version) > 0 {
		fmt.Printf("version %s\n", VERSION)
		os.Exit(0)
	}

	if err != nil {
		flagsErr, ok := err.(*flags.Error)
		if !ok {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s\n", flagsErr.Message)
		os.Exit(0)
	}

	fmt.Printf("target=%v, url=%v\n", opts.Target, opts.URL)

	evCh := make(chan *Event)
	go Notify(opts.URL, evCh)

	for ev := range evCh {
		fmt.Printf("%s %d", string(ev.Data), len(ev.Data))
		if len(ev.Data) <= 2 {
			continue
		}

		contentType, _, _, err := jsonparser.Get(ev.Data, "content-type")
		if err != nil {
			fmt.Printf("Error: no content-type found\n")
			continue
		}

		if opts.Secret != "" {
			signature, _, _, err := jsonparser.Get(ev.Data, "x-hub-signature")
			if err != nil {
				fmt.Printf("Error: no signature found\n")
				continue
			}

			body, _, _, err := jsonparser.Get(ev.Data, "body")
			if err != nil {
				fmt.Printf("Error: no body found\n")
				continue
			}

			if !ValidMAC([]byte(body), hex2bytes(string(signature)), body) {
				fmt.Printf("#v", err)
				fmt.Printf("\nInvalid HMAC\n")
				continue
			}
		}

		_, err = http.Post(opts.Target, string(contentType), bytes.NewBuffer(ev.Data))
		if err != nil {
			fmt.Printf("#v", err)
		}
	}
}
