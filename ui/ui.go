//go:generate npm run build:css
package ui

import "embed"

//go:embed *.html
var Templates embed.FS
