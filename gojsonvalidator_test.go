package main

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/jacobsa/oglematchers"

	"os"
	"flag"
	"io/ioutil"
	"fmt"
	"reflect"
	"path/filepath"
)
func ShouldMatch(actual interface{}, expected ...interface{}) string {
	if (len(expected) != 1) {
		return fmt.Sprintf("This assertion requires exactly %d comparison values (you provided %d).",len(expected))
	}
	value,valueIsString := actual.(string)
	regex,regexIsString := expected[0].(string)

	if !valueIsString || !regexIsString {
		return fmt.Sprintf("Both arguments to this assertion must be strings (you provided %v and %v).",reflect.TypeOf(actual),reflect.TypeOf(expected[0]))
	}

	matcher := MatchesRegexp(regex)

	if err := matcher.Matches(value); err == nil {
		return ""
	} else {
		return fmt.Sprintf("Expected '%v' to match '%v' (but it didn't)!",regex,value)
	}
}
type IORedirecter struct {
	oldRef **os.File
	oldRefValue *os.File
	writer *os.File
	reader *os.File
}
func NewIORedirecter(fd **os.File) (result *IORedirecter) {
	result = &IORedirecter{
		oldRef: fd,
		oldRefValue: *fd,
	}
	var err error
	result.reader,result.writer,err = os.Pipe()
	if err != nil {
		return nil
	}
	return
}
func (ior *IORedirecter) start() {
	*ior.oldRef = ior.writer
}
func (ior *IORedirecter) closeAndGet() string {
	ior.writer.Close()
	bytes,_ := ioutil.ReadAll(ior.reader)
	*ior.oldRef = ior.oldRefValue
	return string(bytes)
}
type TestCase struct {
	schema string
	documents []struct {
		file string
		isValid bool
	 }
}
func TestAll(t *testing.T) {

	Convey("Validating JSON schemas should work",t,func() {
		testMode = true
		verbose = false
		documents = nil
		interactive = false
		schemaFile = ""
		flag.CommandLine = flag.NewFlagSet("test",flag.ContinueOnError)

		stdoutRedirecter := NewIORedirecter(&os.Stdout)
		stderrRedirecter := NewIORedirecter(&os.Stderr)

		Convey("Parsing from STDIN should work",func() {
			stdoutRedirecter.start()
			stderrRedirecter.start()

			os.Args = []string{"", "-s", filepath.Join("tests","products-schema.json"), "-i", "-v" }

			r,w,_ := os.Pipe()
			w.WriteString("[{\"id\":2,\"name\":\"An ice sculpture\",\"price\":12.50,\"tags\":[\"cold\",\"ice\"],\"dimensions\":{\"length\":7.0,\"width\":12.0,\"height\":9.5},\"warehouseLocation\":{\"latitude\":-78.75,\"longitude\":20.4}},{\"id\":3,\"name\":\"A blue mouse\",\"price\":25.50,\"dimensions\":{\"length\":3.1,\"width\":1.0,\"height\":1.0},\"warehouseLocation\":{\"latitude\":54.4,\"longitude\":-32.7}}]")
			w.Close()
			os.Stdin = r

			main()

			So(stdoutRedirecter.closeAndGet(),ShouldMatch,"^JSON is valid")
			So(stderrRedirecter.closeAndGet(),ShouldEqual,"")
			So(exitCode,ShouldEqual,0)
		})

		Convey("Valid products document should work",func() {
			stdoutRedirecter.start()
			stderrRedirecter.start()

			os.Args = []string{"", "-s", filepath.Join("tests","products-schema.json"), "-f", filepath.Join("tests","products.json") }

			main()

			So(stdoutRedirecter.closeAndGet(),ShouldEqual,"")
			So(stderrRedirecter.closeAndGet(),ShouldEqual,"")
			So(exitCode,ShouldEqual,0)

		})

		Convey("Verbose flag should work",func() {
			stdoutRedirecter.start()
			stderrRedirecter.start()

			os.Args = []string{"", "-s", filepath.Join("tests","products-schema.json"), "-f", filepath.Join("tests","products.json"), "-v" }

			main()

			So(stdoutRedirecter.closeAndGet(),ShouldMatch,"(?s)^JSON is valid")
			So(stderrRedirecter.closeAndGet(),ShouldEqual,"")
			So(exitCode,ShouldEqual,0)
		})

		Convey("Invalid products document should not work",func() {
			stdoutRedirecter.start()
			stderrRedirecter.start()

			os.Args = []string{"", "-s", filepath.Join("tests","products-schema.json"), "-f", filepath.Join("tests","products-invalid.json") }

			main()

			So(stdoutRedirecter.closeAndGet(),ShouldEqual,"")
			So(stderrRedirecter.closeAndGet(),ShouldMatch,"(?s)^JSON is invalid.+Invalid type. Expected: number, given: string\n$")
			So(exitCode,ShouldEqual,1)

		})

		Convey("Invalid products schema should not work",func() {
			stdoutRedirecter.start()
			stderrRedirecter.start()

			os.Args = []string{"", "-s", filepath.Join("tests","products-schema.json"), "-f", "bogus.txt" }

			main()

			So(stdoutRedirecter.closeAndGet(),ShouldEqual,"")
			So(stderrRedirecter.closeAndGet(),ShouldMatch,"(?s)^invalid value \"bogus.txt\" for flag -f:")
			So(exitCode,ShouldEqual,0)

		})

	})

}