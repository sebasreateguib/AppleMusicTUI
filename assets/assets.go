// Package assets holds the files built into the binary.
package assets

import _ "embed"

//go:embed am-ascii.txt
var Logo string
