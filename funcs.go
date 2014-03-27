package blueprint

import (
	htemp "html/template"
)

func funcHtmlComment(s string) htemp.HTML {
	return htemp.HTML("<!-- " + s + " -->")
}
