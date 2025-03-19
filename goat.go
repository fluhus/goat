// Command goat generates go source from a given template.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"text/template"
)

var (
	in       = flag.String("i", "", "Path to input template file. If omitted, reads from stdin.")
	out      = flag.String("o", "", "Path to output go file. If omitted, writes to stdout.")
	nh       = flag.Bool("nh", false, "Don't add a header to the output file.")
	nf       = flag.Bool("nf", false, "Don't run gofmt on the result.")
	data     = flag.String("d", "", "JSON-encoded `data` for the template.")
	dataFile = flag.String("df", "", "JSON-encoded `file` with data for the template.")
)

func main() {
	flag.Parse()

	// Parse template data.
	if *data != "" && *dataFile != "" {
		fmt.Fprintln(os.Stderr, "Only one of -d and -df can be used.")
		os.Exit(2)
	}
	var d any
	if *data != "" {
		if err := json.Unmarshal([]byte(*data), &d); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to parse data (-d param):", err)
			os.Exit(2)
		}
	}
	if *dataFile != "" {
		data, err := os.ReadFile(*dataFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not read data file:", err)
			os.Exit(2)
		}
		if err := json.Unmarshal(data, &d); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to parse data file (-df param):", err)
			os.Exit(2)
		}
	}

	// Read template.
	var err error
	var input []byte
	if *in == "" {
		fmt.Fprintln(os.Stderr, "Reading from stdin...")
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(*in)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read input:", err)
		os.Exit(2)
	}

	// Parse template.
	funcs := map[string]interface{}{"slice": makeSlice}
	t, err := template.New("").Funcs(funcs).Parse(string(input))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse template:", err)
		os.Exit(2)
	}

	// Execute template.
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, d)
	if err != nil {
		fmt.Println("Failed to execute template:", err)
		os.Exit(2)
	}
	src := buf.Bytes()

	// Attach header.
	if !*nh {
		from := ""
		if *in != "" {
			from = "from '" + *in + "' "
		}
		header = fmt.Sprintf(header, from)
		src = append([]byte(header), src...)
	}

	// Run gofmt.
	if !*nf {
		src, err = format.Source(src)
		if err != nil {
			fmt.Println("Failed to gofmt the resulting source:", err)
			os.Exit(2)
		}
	}

	// Write output.
	if *out == "" {
		fmt.Print(string(src))
	} else {
		err = os.WriteFile(*out, src, 0644)
		if err != nil {
			fmt.Println("Failed to write output:", err)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "Wrote to:", *out)
	}
}

func makeSlice(a ...interface{}) []interface{} {
	return a
}

var header = `// ***** DO NOT EDIT THIS FILE MANUALLY. *****
//
// This file was auto-generated %vusing goat.
//
// goat: https://www.github.com/fluhus/goat

`
