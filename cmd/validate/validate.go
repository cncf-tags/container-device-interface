/*
 * Original code from: https://github.com/opencontainers/runtime-spec/blob/643c1429d905bba70fe977bae274f367ad101e73/schema/validate.go
 * Changes:
 *  - Output errors to stderr
 *  - Refactored to use package-internal validation library
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"tags.cncf.io/container-device-interface/schema"
)

const usage = `Validate is used to check document with specified schema.
You can use validate in following ways:

   1.specify document file as an argument
      validate --schema <schema.json> <document.json>

   2.pass document content through a pipe
      cat <document.json> | validate --schema <schema.json>

   3.input document content manually, ended with ctrl+d(or your self-defined EOF keys)
      validate --schema <schema.json>
      [INPUT DOCUMENT CONTENT HERE]
`

func main() {
	var (
		schemaFile string
		docFile    string
		docData    []byte
		err        error
		exitCode   int
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", usage)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&schemaFile, "schema", "builtin", "JSON Schema to validate against")
	flag.Parse()

	if schemaFile != "" {
		scm, err := schema.Load(schemaFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load schema %s: %v\n", schemaFile, err)
			os.Exit(1)
		}
		schema.Set(scm)
		fmt.Printf("Validating against JSON schema %s...\n", schemaFile)
	} else {
		fmt.Printf("Validating against builtin JSON schema...\n")
	}

	docs := flag.Args()
	if len(docs) == 0 {
		docs = []string{"-"}
	}

	for _, docFile = range docs {
		if docFile == "" || docFile == "-" {
			docFile = "<stdin>"
			docData, err = io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to read document data from stdin: %v\n", err)
				os.Exit(1)
			}
			err = schema.ValidateData(docData)
		} else {
			err = schema.ValidateFile(docFile)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: validation failed:\n    %v\n", docFile, err)
			exitCode = 1
		} else {
			fmt.Printf("%s: document is valid.\n", docFile)
		}
	}

	os.Exit(exitCode)
}
