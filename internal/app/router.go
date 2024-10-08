package app

import (
	"fmt"
	"intermark/internal/database"
	"intermark/internal/utils"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/Data-Corruption/blog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var Templates *template.Template

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		blog.Debugf("Started %s %s", r.Method, r.URL.Path)
		start := time.Now()
		next.ServeHTTP(ww, r)
		blog.Debugf("Completed %s in %v with status %d", r.URL.Path, time.Since(start), ww.Status())
	})
}

func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (utils.Config.Server.CacheMaxAge <= 0) || (strings.HasSuffix(r.URL.Path, ".css")) {
			// If max-age is 0 or negative, disable caching
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		} else {
			// Set Cache-Control header with the specified max-age
			cacheControl := fmt.Sprintf("public, max-age=%d", utils.Config.Server.CacheMaxAge)
			w.Header().Set("Cache-Control", cacheControl)
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter creates and returns a new Chi router.
func NewRouter(usingTLS *bool) *chi.Mux {
	r := chi.NewRouter()

	// load the templates
	var err error
	if Templates, err = template.ParseGlob("data/templates/*.html"); err != nil {
		blog.Fatalf(1, time.Second*3, "Error parsing templates: %s", err)
	}

	// add middleware
	r.Use(logMiddleware)

	// cached routes
	r.Group(func(r chi.Router) {
		r.Use(cacheControlMiddleware)
		r.Get("/css/*", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "data"+r.URL.Path) })
		// If config asset dir is "assets", your src vars will look like "/assets/example.png"
		// Should mean they still path correctly in the content repo and when served in the site.
		r.Get(fmt.Sprintf("/%s/*", utils.Config.ContentRepo.AssetsDir), func(w http.ResponseWriter, r *http.Request) {
			database.AssetsMutex.RLock()
			defer database.AssetsMutex.RUnlock()
			http.ServeFile(w, r, "data/assets"+r.URL.Path)
		})
		r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			database.AssetsMutex.RLock()
			defer database.AssetsMutex.RUnlock()
			http.ServeFile(w, r, fmt.Sprintf("data/assets/%s/logo.svg", utils.Config.ContentRepo.AssetsDir))
		})
	})

	// helper func for serving pages
	var servePage = func(w http.ResponseWriter, id string, template string) {
		html, err := database.GetHTML(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{"Title": utils.Config.Title, "Layout": database.GetLayout(), "Content": html, "Hamburger": true, "Edit": false}
		if err := Templates.ExecuteTemplate(w, template, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// pages
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		servePage(w, database.GetLayout().Landing.ID, "landing.html")
	})
	r.Get("/page", func(w http.ResponseWriter, r *http.Request) {
		servePage(w, r.URL.Query().Get("id"), "page.html")
	})

	// edit
	r.Get("/edit", GetEditLogin())
	r.Post("/edit", PostEditLogin(usingTLS))
	r.Group(func(r chi.Router) {
		r.Use(EditAuthMiddleware)
		r.Post("/edit/new-sidebar-item", PostEditNewSidebarItem())
		r.Post("/edit/new-footer-item", PostEditNewFooterItem())
		r.Post("/edit/update-sandbox", PostEditUpdateSandbox())
		r.Post("/edit/update-content", PostEditUpdateContent())
		r.Post("/edit/save", PostEditSave())
		r.Post("/edit/exit", PostEditExit())
	})

	// update from content repo action
	r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		if string(body) != utils.Config.UpdateToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err := database.Update(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found"))
	})

	return r
}
