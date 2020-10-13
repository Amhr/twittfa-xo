package template

import (
	"OnlineXO/config"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplate() *Template {
	return &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

func TemplateData(p echo.Map) echo.Map {
	if p == nil {
		p = echo.Map{}
	}
	p["baseurl"] = config.FULL_URL
	p["base_url"] = config.FULL_URL
	p["site_title"] = "TwittFa"
	return p
}
