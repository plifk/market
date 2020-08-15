package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/plifk/market/internal/config"
	"github.com/plifk/market/internal/services"
	"github.com/plifk/market/internal/validator"
)

// HTMLResponse to send to the client.
type HTMLResponse struct {
	Template      string
	Title         string
	CanonicalLink string
	Breadcrumb    []Breadcrumb
	Content       interface{}
	SkipLayout    bool

	Params *HTMLResponseParams
}

func (f *Frontend) prepareTemplates(dir string) (*template.Template, error) {
	t := template.New("x") // The template name is not being used anywhere.
	t = t.Funcs(basicTemplateFuncs)
	t = t.Funcs(template.FuncMap{
		"img": func(path string, width int, alt string) (template.HTML, error) {
			return f.Modules.Images.HTML(path, alt, services.ThumbnailParams{Width: width})
		},
	})
	var err error
	t, err = t.ParseGlob(dir + "/**.html")
	if err != nil {
		return nil, fmt.Errorf("cannot parse HTML templates: %w", err)
	}
	return t, nil
}

// HTMLResponseParams injected into the template object.
type HTMLResponseParams struct {
	Settings  *config.Settings
	CSRFField template.HTML
	Request   *http.Request
	User      *services.User
}

// Respond HTML to the browser.
//
// NOTE(henvic): Regarding why we need to use a second template execution when printing with the layout, please see:
// https://groups.google.com/d/msg/golang-nuts/PRLloHrJrDU/lDYA6Pq_l8QJ
// https://stackoverflow.com/questions/28830543/how-to-use-a-field-of-struct-or-variable-value-as-template-name/28831138#28831138
func (f *Frontend) Respond(w http.ResponseWriter, r *http.Request, resp *HTMLResponse) {
	settings := f.Modules.Settings
	t, err := f.prepareTemplates(settings.TemplatesDirectory) // TODO(henvic): cache in production.
	if err != nil {
		log.Printf("cannot parse template files: %v\n", err)
		http.Error(w, fmt.Sprintf("%v: template error", http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError)
		return
	}

	resp.Params = &HTMLResponseParams{
		Settings:  &f.Modules.Settings,
		CSRFField: csrfField(r),
		Request:   r,
		User:      services.UserFromRequest(r),
	}

	var writer io.Writer = w // If response doesn't use layout, do not buffer.
	var buf bytes.Buffer
	if !resp.SkipLayout {
		writer = &buf
	}
	if err := t.ExecuteTemplate(writer, resp.Template, resp); err != nil {
		log.Printf("cannot print template %q: %v\n", resp.Template, err)
		http.Error(w, fmt.Sprintf("%v: %q template execution error", http.StatusText(http.StatusInternalServerError), resp.Template), http.StatusInternalServerError)
		return
	}
	if !resp.SkipLayout {
		wrapLayout(w, t, &buf, resp)
		buf.WriteTo(w)
	}
}

func wrapLayout(w http.ResponseWriter, t *template.Template, buf *bytes.Buffer, data *HTMLResponse) {
	// Copy the bytes of the response from the buffer to the Body field.
	layout := struct {
		*HTMLResponse
		Body template.HTML
	}{
		HTMLResponse: data,
		Body:         template.HTML(buf.Bytes()),
	}
	// Clean up the buffer after copying.
	buf.Reset()
	if err := t.ExecuteTemplate(buf, "layout", &layout); err != nil {
		log.Printf("cannot print template %q: %v\n", "layout", err)
		http.Error(w, fmt.Sprintf("%v: %q template execution error", http.StatusText(http.StatusInternalServerError), "layout"), http.StatusInternalServerError)
		return
	}
}

var basicTemplateFuncs = template.FuncMap{
	"json": func(v interface{}) string {
		a, _ := json.Marshal(v)
		return string(a)
	},
	"split":      strings.Split,
	"join":       strings.Join,
	"title":      strings.Title,
	"lower":      strings.ToLower,
	"upper":      strings.ToUpper,
	"formErrors": validator.TemplateErrors,
}

func csrfField(r *http.Request) template.HTML {
	var token string
	if r != nil {
		token = services.CSRFToken(r)
	}
	return template.HTML(`<input type="hidden" name="csrf_token" value="` + html.EscapeString(token) + `">`)
}

// Breadcrumb shows where the user is currently.
type Breadcrumb struct {
	Text   string
	Link   string
	Active bool
}
