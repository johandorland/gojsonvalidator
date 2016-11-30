# gojsonvalidator

## Description
gojsonvalidator is a command-line tool based on [gojsonschema](https://github.com/xeipuuv/gojsonschema) to validate JSON documents using [JSON schema](http://json-schema.org/).


```
Usage of gojsonvalidator:
  -f value
    	One or more document files to validate.
  -i	Parse a single JSON document from STDIN. (Works in conjunction with -f)
  -s string
    	Schema file to validate documents with. (default "schema.json")
  -v	Print verbose output about all files.

```


## Installation
To install **gojsonvalidator** use ```go get ```
```
$ go get github.com/johandorland/gojsonvalidator
```

## Dependencies
Dependencies are managed using [Glide](https://github.com/Masterminds/glide)
* [github.com/xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema) for providing JSON schema validation
* [github.com/smartystreets/goconvey](https://github.com/smartystreets/goconvey) for testing.

Note that if you want to use GoConvey's web interface while testing, install it using ```go get github.com/smartystreets/goconvey``` as it does not support installation in the vendor directory.
