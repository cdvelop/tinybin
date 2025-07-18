module github.com/cdvelop/tinybin

go 1.24.4

require github.com/cdvelop/tinyreflect v0.0.27

require (
	github.com/cdvelop/tinystring v0.1.33
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/cdvelop/tinyreflect v0.0.27 => ../tinyreflect
