package main

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type Page struct {
    Title string
    Body []byte
}

func (p *Page) save() error {
    filename := "data/" + p.Title + ".txt"
    return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    filename := "data/" + title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))
var pageLinkRegex = regexp.MustCompile(`\[([^\[\]]+)\]`)

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	// Replace instances of [PageName] with link to view page
	p.Body = []byte(pageLinkRegex.ReplaceAllStringFunc(string(p.Body), func(match string) string {
		pageName := strings.Trim(match, "[]") // Extract PageName
		return "<a href=\"/view/" + pageName + "\">" + pageName + "</a>"
	}))
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}