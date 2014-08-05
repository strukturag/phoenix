# Phoenix

[![GoDoc](https://godoc.org/github.com/strukturag/phoenix?status.svg)](https://godoc.org/github.com/strukturag/phoenix)
[![Build Status](https://travis-ci.org/strukturag/phoenix.png?branch=master)](https://travis-ci.org/strukturag/phoenix)

## Introduction
Package phoenix provides runtime support for long running server processes.

In particular, it provides standardized mechanisms for handling logging,
configuration, and HTTP server startup, as well as profiling support.

Additionally, lifecycle management facilities for application services
which should respond to server state changes are provided, as are a
full suite of signal handling functionality, including configuration
reload on SIGHUP.

## Usage

Import into your workspace via `go get`:

```bash
go get github.com/strukturag/phoenix
```

And then use in your application as follows:

```go
package main

import (
    "flag"
	"fmt"
    "net/http"
    "os"

    "github.com/strukturag/phoenix"
)

var version = "unreleased"
var defaultConfig = "./server.conf"

func boot() error {
	configPath := flag.String("c", defaultConfig, "Configuration file.")
	logPath := flag.String("l", "", "Log file, defaults to stderr.")
	showVersion := flag.Bool("v", false, "Display version number and exit.")
    memprofile := flag.String("memprofile", "", "Write memory profile to this file.")
    cpuprofile := flag.String("cpuprofile", "", "Write cpu profile to file.")
	showHelp := flag.Bool("h", false, "Show this usage information and exit.")
	flag.Parse()

	if *showHelp {
		flag.Usage()
		return nil
	} else if *showVersion {
		fmt.Printf("Version %s\n", version)
		return nil
	}

	return phoenix.NewServer("myapp", version).
		Config(configPath).
		Log(logPath).
		CpuProfile(cpuprofile).
		MemProfile(memprofile).
		Run(func(runtime phoenix.Runtime) error {
            runtime.DefaultHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
                if _, err := w.Write([]byte("Hello again, phoenix!\n")); err != nil {
					runtime.Printf("Failed to write response: %v", err)
				}
            }))

            return runtime.Start()
        })
}

func main() {
   if err := boot(); err != nil {
       os.Exit(-1)
   }
}
```

## License

This package is licensed by struktur AG under the 3-clause BSD license,
see LICENSE for more details.
