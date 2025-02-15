package main

import (
	"bytes"
	"encoding/xml"
)

func xmlEscape(s string) string {
	var b bytes.Buffer
	xml.EscapeText(&b, []byte(s))
	return b.String()
}
