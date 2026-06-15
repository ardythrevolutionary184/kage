package pack

import (
	"errors"
	"net/http"
	"strings"

	"github.com/tamnd/kage/zim"
)

// Handler serves a ZIM archive over HTTP. "/" redirects to the archive's main
// page; "/a/b.png" maps to the C/a/b.png content entry. Because the saved HTML's
// links are mirror-relative paths, and those are exactly the C urls, a click in a
// served page hits the right entry with no rewriting. A miss is a plain 404.
func Handler(r *zim.Reader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		p := strings.TrimPrefix(req.URL.Path, "/")
		if p == "" {
			// The main page's saved HTML carries mirror-relative asset URLs
			// (../_kage/...) computed for its own nested location, so serving its
			// bytes at "/" would resolve them against the wrong base and 404 the
			// page's CSS and images. Redirect to the page's canonical content path
			// instead, the way the archive's W/mainPage redirect does, so the
			// browser resolves those relative URLs correctly.
			if ns, url, ok := r.MainPageRef(); ok && ns == zim.NamespaceContent {
				http.Redirect(w, req, "/"+url, http.StatusFound)
				return
			}
			blob, err := r.MainPage()
			if err != nil {
				http.NotFound(w, req)
				return
			}
			serveBlob(w, blob)
			return
		}
		blob, err := r.Get(zim.NamespaceContent, p)
		if errors.Is(err, zim.ErrNotFound) {
			http.NotFound(w, req)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		serveBlob(w, blob)
	})
}

func serveBlob(w http.ResponseWriter, b zim.Blob) {
	if b.MimeType != "" {
		w.Header().Set("Content-Type", b.MimeType)
	}
	_, _ = w.Write(b.Data)
}
