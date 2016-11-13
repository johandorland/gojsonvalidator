package main

import (
	"fmt"
	"flag"
	"os"
	"github.com/xeipuuv/gojsonschema"
	"sync"
	"path/filepath"
	"io/ioutil"
	"io"
)

var interactive bool
var verbose bool
var schemaFile string
var testMode bool = false
var exitCode int

type Documents []gojsonschema.JSONLoader

type ResultTuple struct {
 	*gojsonschema.Result
	err error
	source string
}

func (d Documents) String() string {
	return "";
}
func (d*Documents) Set(value string) error {
	 if _,err := os.Stat(value); err != nil {
		 return err
	 }
	absolutePath,err := filepath.Abs(value)
	if err != nil {
		return err
	}
	*d = append(*d,gojsonschema.NewReferenceLoader("file://"+absolutePath))
	return nil
}
var documents Documents

func main() {
	parseArguments()

	absSchemaFile,err := filepath.Abs(schemaFile)
	if err != nil {
		fmt.Fprintln(os.Stderr,"An error occurred while loading "+schemaFile +" from the file system.")
		exit(1)
	}

	rawSchema := gojsonschema.NewReferenceLoader("file://"+absSchemaFile)
	schema,err := gojsonschema.NewSchema(rawSchema)

	if err != nil {
		fmt.Fprintln(os.Stderr,"An error occured while passing the following schema file: "+schemaFile)
		fmt.Fprintln(os.Stderr,err.Error())
		exit(1)
	}
	if interactive {
		getDocumentFromReader(os.Stdin)
	}

	results := validateDocuments(schema,documents)

	exit(printResults(results))

}
func exit(code int) {
	if !testMode {
		os.Exit(code)
	} else {
		exitCode = code
	}
}
func parseArguments() {
	flag.BoolVar(&verbose, "v", false, "Print verbose output about all files.")
	flag.BoolVar(&interactive,"i",false,"Parse a single JSON document from STDIN. (Works in conjunction with -f)")
	flag.StringVar(&schemaFile, "s", "schema.json", "Schema file to validate documents with.")
	flag.Var(&documents,"f","One or more document files to validate.")
	flag.Parse()
}
func getDocumentFromReader(input io.Reader) {
	if interactive {
		if bytes,err := ioutil.ReadAll(input); err == nil {
			documents = append(documents,gojsonschema.NewBytesLoader(bytes))
		}
	}
}
func printResults(results []ResultTuple) (exitCode int){
	prependGuard := func(line string) string {
		if line == "" {
			return line
		}
		return "| " + line
	}
	for _, result := range results {
		if result.err != nil {
			fmt.Fprintf(os.Stderr,"An error occured %s %s\n", prependGuard(result.source),result.err.Error())
			exitCode = 1
		} else if result.Valid() {
			if verbose {
				fmt.Fprintf(os.Stdout,"JSON is valid %s\n", prependGuard(result.source))
			}
		} else {
			fmt.Fprintf(os.Stderr,"JSON is invalid %s\n", prependGuard(result.source))
			for _,err := range result.Errors() {
				fmt.Fprintf(os.Stderr,"- %s\n",err)
			}
			exitCode = 1
		}
	}
	return
}
func validateDocuments(schema *gojsonschema.Schema,documents []gojsonschema.JSONLoader) []ResultTuple {
	results := make([]ResultTuple, len(documents))

	var done sync.WaitGroup
	done.Add(len(documents))

	for index, document := range documents {

		go func(i int,d gojsonschema.JSONLoader) {
			validation, err := schema.Validate(d)

			var result ResultTuple
			result.Result = validation
			result.err = err

			jref,_ := d.JsonReference()
			if jref.String() != "" {
				result.source = jref.String()
			}

			results[i] = result
			done.Done()
		}(index,document)
	}
	done.Wait()

	return results
}