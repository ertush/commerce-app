package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type RootHandler struct {
	// Add fields here
}

func NewRootHandler() *RootHandler {
	return &RootHandler{}
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	md_str, err := os.ReadFile("docs/REST_API_DOCS.md")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := mdToHTML(md_str)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w,
		`
		<body style="
					margin: 0;
				    padding: 40px 0px 40px 0px;
				    font-family: Arial, sans-serif;
				    padding: 20px;
				    background-color: #333;
				    color: #f5f5f5;"
				>
				%s
		</body>
		`,
		string(html))

}
