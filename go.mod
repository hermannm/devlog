module hermannm.dev/devlog

go 1.26.0

require (
	github.com/neilotoole/jsoncolor v0.10.1
	golang.org/x/sys v0.48.0
	golang.org/x/term v0.46.0
)

// Test dependencies
require github.com/stretchr/testify v1.12.1

// Transitive test dependencies
require go.yaml.in/yaml/v3 v3.0.5 // indirect
