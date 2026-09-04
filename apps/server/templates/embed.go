// Package templates provides the shared layout for email paths that render
// outside the file-backed templated mailer, including Maya's reply processor.
package templates

import _ "embed"

//go:embed layouts/base.html
var BaseLayout string
