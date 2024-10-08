package app

import (
	"bytes"
	"encoding/json"
	"html/template"
	"intermark/internal/database"
	"intermark/internal/utils"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Data-Corruption/blog"
)

type genericRequest struct {
	Token string          `json:"token"`
	Data  json.RawMessage `json:"data"`
}
type saveReq struct {
	Token string `json:"token"`
	Data  struct {
		Layout database.Layout `json:"layout"`
	} `json:"data"`
}
type newItemReq struct {
	Token string `json:"token"`
	Data  struct {
		Type string `json:"type"`
	} `json:"data"`
}
type sandboxReq struct {
	Token string `json:"token"`
	Data  struct {
		SandboxMD string `json:"sandbox_md"`
	} `json:"data"`
}

var (
	rateLimitMutex   sync.Mutex
	editSessionMutex sync.Mutex
	editSessionToken string = "" // empty = no session
)

func EditAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		editSessionMutex.Lock()
		// if no edit session token, redirect to login
		if editSessionToken == "" {
			editSessionMutex.Unlock()
			http.Redirect(w, r, "/edit", http.StatusSeeOther)
			return
		}
		// read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			editSessionMutex.Unlock()
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		r.Body.Close()
		// create a new reader from the bytes
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// parse the body
		var req genericRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			editSessionMutex.Unlock()
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		// get sessionToken cookie
		cookie, err := r.Cookie("sessionToken")
		if err != nil {
			editSessionMutex.Unlock()
			http.Error(w, "No session token", http.StatusUnauthorized)
			return
		}
		// check if token or cookie is invalid
		if (req.Token != editSessionToken) || (cookie.Value != editSessionToken) {
			editSessionMutex.Unlock()
			http.Error(w, "Invalid token or cookie", http.StatusUnauthorized)
			return
		}
		editSessionMutex.Unlock()
		next.ServeHTTP(w, r)
	})
}

func GetEditLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{"Title": utils.Config.Title, "Layout": database.GetLayout(), "Hamburger": false, "Edit": false}
		if err := Templates.ExecuteTemplate(w, "editLogin.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func PostEditLogin(usingTLS *bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get password from json body
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		password := r.FormValue("password")
		// rate limit
		rateLimitMutex.Lock()
		go func() {
			time.Sleep(12 * time.Second)
			rateLimitMutex.Unlock()
		}()
		// check password
		if password != utils.Config.EditPassword {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}
		// gen new token
		editSessionMutex.Lock()
		defer editSessionMutex.Unlock()
		newToken, err := utils.GenRandomString(32)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		editSessionToken = newToken
		// set cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "sessionToken",
			Value:    newToken,
			Path:     "/edit",
			Secure:   *usingTLS,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		// get layout
		var layout *database.Layout
		if layout, err = database.GetLayoutDB(); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// get meta data for all pages
		var metaDatas []database.ContentMeta
		if metaDatas, err = database.GetMeta(); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		var metaBytes []byte
		if metaBytes, err = json.Marshal(metaDatas); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// init sandbox
		var sandboxHTML string
		if sandboxHTML, err = database.UpdateSandbox(database.DEFAULT_SANDBOX_MD); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// render edit page, embed token
		data := map[string]interface{}{
			"Token":            newToken,
			"Title":            utils.Config.Title,
			"Layout":           *layout,
			"PageMetaDataJSON": template.JS(metaBytes),
			"UpdateTimeout":    utils.Config.UpdateTimeout * 1000,
			"Hamburger":        true,
			"Edit":             true,
			"SandboxMD":        database.DEFAULT_SANDBOX_MD,
			"SandboxHTML":      template.HTML(sandboxHTML),
		}
		if err := Templates.ExecuteTemplate(w, "edit.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func newItemHelper(w http.ResponseWriter, r *http.Request, templateName string) {
	var newItem newItemReq
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var data map[string]interface{}
	if templateName == "edit_sidebar_item" {
		data = map[string]interface{}{"Type": newItem.Data.Type, "Name": "New " + newItem.Data.Type, "Meta": database.ContentMeta{}}
	} else {
		data = map[string]interface{}{"Type": newItem.Data.Type, "Name": "New " + newItem.Data.Type, "Meta": database.ContentMeta{}}
	}
	if err := Templates.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Exposes the edit_sidebar_item template to the front end. Has no server side effect.
func PostEditNewSidebarItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newItemHelper(w, r, "edit_sidebar_item")
	}
}

// Exposes the edit_footer_item template to the front end. Has no server side effect.
func PostEditNewFooterItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newItemHelper(w, r, "edit_footer_item")
	}
}

func PostEditUpdateSandbox() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sandboxReq sandboxReq
		if err := json.NewDecoder(r.Body).Decode(&sandboxReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if newHTML, err := database.UpdateSandbox(sandboxReq.Data.SandboxMD); err != nil {
			blog.Debugf("Error updating sandbox: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write([]byte(newHTML))
		}
	}
}

// PostEditUpdateContent
func PostEditUpdateContent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		if err = database.Update(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// get new meta data for all pages and return it
		var metaDatas []database.ContentMeta
		if metaDatas, err = database.GetMeta(); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		var metaBytes []byte
		if metaBytes, err = json.Marshal(metaDatas); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Write(metaBytes)
	}
}

// PostEditSave
func PostEditSave() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var saveReq saveReq
		if err := json.NewDecoder(r.Body).Decode(&saveReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		blog.Debugf("New Layout: %v", saveReq.Data.Layout)
		database.UpdateMutex.Lock() // avoid writing a new layout while the database is updating
		defer database.UpdateMutex.Unlock()
		database.SetLayout(&saveReq.Data.Layout)
	}
}

// PostEditExit
func PostEditExit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		blog.Debug("Exiting edit session")
		editSessionMutex.Lock()
		editSessionToken = ""
		editSessionMutex.Unlock()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
