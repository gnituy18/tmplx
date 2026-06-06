package main

import "testing"

// The script block can import packages and use them in expressions; the import
// is carried into the generated code.
func TestScriptImportUsable(t *testing.T) {
	ctx := newPage(t, "/importer")
	waitText(t, ctx, `[data-test="upper"]`, "HELLO") // strings.ToUpper("hello")
}
