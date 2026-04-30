module hermannm.dev/devlog

go 1.25.0

require (
	github.com/neilotoole/jsoncolor v0.7.2
	golang.org/x/sys v0.43.0
	golang.org/x/term v0.42.0
)

// Testing dependencies
require github.com/stretchr/testify v1.11.1

// Transitive testing dependencies
require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
