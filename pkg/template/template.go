// +build dev

package template

import "net/http"

var Assets http.FileSystem = http.Dir("tmpl")
