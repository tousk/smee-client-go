package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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

	evCh := make(chan *Event)
	go Notify(opts.URL, evCh)

	for ev := range evCh {
		if len(ev.Data) <= 2 {
			continue
		}

		contentType, _, _, err := jsonparser.Get(ev.Data, "content-type")
		if err != nil {
			fmt.Printf("Error: no content-type found\n")
			continue
		}

		github_event, _, _, err := jsonparser.Get(ev.Data, "x-github-event")
		if err != nil {
			fmt.Printf("Error: no x-github-event found\n")
			continue
		}

		github_delivery, _, _, err := jsonparser.Get(ev.Data, "x-github-delivery")
		if err != nil {
			fmt.Printf("Error: no x-github-delivery found\n")
			continue
		}

		body, _, _, err := jsonparser.Get(ev.Data, "body")
		if err != nil {
			fmt.Printf("Error: no body found\n")
			continue
		}

		if opts.Secret != "" {
			signature, _, _, err := jsonparser.Get(ev.Data, "x-hub-signature")
			if err != nil {
				fmt.Printf("Error: no signature found\n")
				continue
			}
			if string(signature[:5]) != "sha1=" {
				fmt.Printf("Warning: Skipping checking. signature is not SHA1: %s\n", signature)
				continue
			} else {
				if !ValidMAC([]byte(body), hex2bytes(string(signature[5:])), []byte(opts.Secret)) {
					fmt.Printf("\nError: Invalid HMAC\n")
					continue
				}
			}
		}

		fmt.Printf("Received %s", string(ev.Data))

		req, err := http.NewRequest("POST", opts.Target, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", string(contentType))
		req.Header.Set("x-github-event", string(github_event))
		req.Header.Set("x-github-delivery", string(github_delivery))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		rspbody, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(rspbody))
	}
}
