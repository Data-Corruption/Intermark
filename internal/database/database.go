package database

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"intermark/internal/files"
	"intermark/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Data-Corruption/blog"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ==== Variables =============================================================

const (
	DB_PATH           = "./data/data.db"
	CONTENT_REPO_PATH = "./data/content"
	CONTENT_HTML_PATH = "./data/html"
	MISSING_FILE      = "MISSING_FILE"
)

var (
	DB *gorm.DB = nil
	// Value type is Layout
	layoutCache   = atomic.Value{}
	UpdateMutex   = sync.Mutex{}
	AssetsMutex   = sync.RWMutex{}
	tailwindMutex = sync.Mutex{}
	sandboxMutex  = sync.Mutex{}
	//go:embed defaultSandbox.md
	DEFAULT_SANDBOX_MD string
	sandboxMD          = DEFAULT_SANDBOX_MD
	sandBoxHTML        = ""
)

// ==== Types =================================================================

type Layout struct {
	Sidebar []SidebarItem `json:"Sidebar"`
	Footer  []FooterItem  `json:"Footer"`
	Landing ContentMeta   `json:"Landing"`
}

type SidebarItem struct {
	Name     string        `json:"Name"`
	Type     string        `json:"Type"`     // "file", "folder", or "divider"
	Meta     ContentMeta   `json:"Meta"`     // only for file
	Contents []SidebarItem `json:"Contents"` // only for folder
}

type FooterItem struct {
	Name string      `json:"Name"`
	Type string      `json:"Type"` // "footer-text", "footer-link", or "footer-file"
	Meta ContentMeta `json:"Meta"` // only for file
	Link string      `json:"Link"` // only for link
}

type ContentMeta struct {
	ID      string `json:"ID" gorm:"primaryKey"`
	Commit  string `json:"Commit"`
	RelPath string `json:"RelPath"`
}

// ==== Model Definitions =====================================================

// LayoutModel. Aside from the ID, this is just a marshalled JSON of the Layout struct.
type LayoutModel struct {
	ID      uint `gorm:"primaryKey"`
	Sidebar string
	Footer  string
	Landing string
	Missing string
}

type ContentModel struct {
	ContentMeta
	HTML string
	MD   string
}

type AssetModel struct {
	ID     string `json:"ID" gorm:"primaryKey"` // rel path
	Commit string `json:"Commit"`
}

// ==== Public Functions ======================================================

// Init initializes the database connection and migrates the schemas.
// Returns immediately if the database is already initialized.
func Init() {
	// if already initialized, return
	if DB != nil {
		return
	}

	// open the database
	db, err := gorm.Open(sqlite.Open(DB_PATH), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent), // logger.Silent or logger.Info
		PrepareStmt: true,
	})
	if err != nil {
		blog.Fatalf(1, time.Second*3, "failed to connect database: %v", err)
	}

	// migrate the schemas
	if err = db.AutoMigrate(&LayoutModel{}, &ContentModel{}, &AssetModel{}); err != nil {
		blog.Fatalf(1, time.Second*3, "failed to migrate database: %v", err)
	}

	// set the global database variable
	DB = db

	// calculate the default sandbox html
	sandBoxHTML, err = utils.MdToHTML(sandboxMD)
	if err != nil {
		blog.Fatalf(1, time.Second*3, "failed to convert sandbox markdown to html: %v", err)
	}

	if utils.DebugMode {
		if err := RunTailwind(false); err != nil {
			blog.Fatalf(1, time.Second*3, "Failed to run tailwind: %v", err)
		}
	}

	// cache the layout
	var layout *Layout
	if layout, err = GetLayoutDB(); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			blog.Fatalf(1, time.Second*3, "failed to cache layout: %v", err)
		} else {
			// create the default layout
			defaultLayout := Layout{
				Sidebar: []SidebarItem{},
				Footer: []FooterItem{
					{Name: "© /year Intermark", Type: "footer-text"},
					{Name: "Example Link", Type: "footer-link", Link: "https://google.com"},
				},
				Landing: ContentMeta{},
			}
			if err := SetLayout(&defaultLayout); err != nil {
				blog.Fatalf(1, time.Second*3, "failed to initialize layout: %v", err)
			}
		}
	} else {
		layoutCache.Store(*layout)
	}
}

// Close closes the database connection.
func Close() {
	if DB != nil {
		db, err := DB.DB()
		if err != nil {
			panic("failed to get database connection")
		}
		db.Close()
	}
}

func GetLayout() Layout {
	layout := layoutCache.Load().(Layout)
	for i, footerItem := range layout.Footer {
		if footerItem.Type == "footer-text" {
			// replace "/year" substring with the current year, e.g. "2012"
			layout.Footer[i].Name = strings.ReplaceAll(footerItem.Name, "/year", fmt.Sprint(time.Now().Year()))
		}
	}
	return layout
}

// GetLayoutDB retrieves the Layout from the db. For frequent access, use LayoutCache.Load()
func GetLayoutDB() (*Layout, error) {
	var result Layout
	var layout LayoutModel
	if err := DB.First(&layout).Error; err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(layout.Sidebar), &result.Sidebar); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(layout.Footer), &result.Footer); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(layout.Landing), &result.Landing); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetLayout sets the Layout in the db and updates the cache.
func SetLayout(layout *Layout) error {
	blog.Debugf("Setting layout: %v", layout)
	// marshal the layout
	mS, err := json.Marshal(layout.Sidebar)
	if err != nil {
		return err
	}
	blog.Debugf("Marshalled Sidebar: %s", mS)
	mF, err := json.Marshal(layout.Footer)
	if err != nil {
		return err
	}
	blog.Debugf("Marshalled Footer: %s", mF)
	mL, err := json.Marshal(layout.Landing)
	if err != nil {
		return err
	}
	blog.Debugf("Marshalled Landing: %s", mL)
	// save the layout and cache it
	if err := DB.Save(&LayoutModel{ID: 1, Sidebar: string(mS), Footer: string(mF), Landing: string(mL)}).Error; err != nil {
		return err
	}
	layoutCache.Store(*layout)
	return nil
}

// GetHTML retrieves the HTML content with the given id.
func GetHTML(id string) (string, error) {
	var content ContentModel
	if err := DB.Select("html").Where("id = ?", id).First(&content).Error; err != nil {
		return `<h2>Oops... Page Not Found!</h2>`, utils.Ternary(errors.Is(err, gorm.ErrRecordNotFound), nil, err)
	}
	return content.HTML, nil
}

// Update updates the content for the site. Returned errors are generic and safe to display.
func Update() error {
	if utils.Config.ContentRepo.URL == "" {
		blog.Error("ContentRepo URL is empty")
		return errors.New("content repository URL in config is empty")
	}

	UpdateMutex.Lock()
	defer UpdateMutex.Unlock()

	var commit string

	// clone or update the content repository
	if exists, err := files.Exists(filepath.Join(CONTENT_REPO_PATH, ".git")); err != nil {
		blog.Errorf("Error checking if a .git directory already exists: %v", err)
		return errors.New("error checking if a .git directory already exists")
	} else if !exists {
		if commit, err = utils.GitClone(utils.Config.ContentRepo.URL, CONTENT_REPO_PATH); err != nil {
			blog.Errorf("Error cloning the content repository: %v", err)
			return errors.New("error cloning the content repository")
		}
		blog.Debugf(`Cloned: '%s', commit: '%s'`, utils.Config.ContentRepo, commit)
	} else {
		if commit, err = utils.GitReset(CONTENT_REPO_PATH); err != nil {
			blog.Errorf("Error resetting the content repository: %v", err)
			return errors.New("error resetting the content repository")
		}
		blog.Debugf(`Reset: '%s', commit: '%s'`, utils.Config.ContentRepo, commit)
	}

	// load the new meta data for all pages
	var err error
	var newMetaDatas []ContentMeta
	if newMetaDatas, err = loadIDs(CONTENT_REPO_PATH); err != nil {
		blog.Errorf("Error loading ids: %v", err)
		return errors.New("error loading ids")
	}
	if len(newMetaDatas) == 0 {
		blog.Warnf("No ids found, skipping update")
		return nil
	}

	// get the old meta data, copy the commits from old to new, and clean the html directory
	var oldMetaDatas []ContentMeta
	if oldMetaDatas, err = GetMeta(); err != nil {
		blog.Errorf("Error getting current meta data: %v", err)
		return errors.New("error getting current content meta data from the database")
	}
	copyCommits(oldMetaDatas, newMetaDatas)

	// update the content
	for _, metaData := range newMetaDatas {
		if err = updateContent(CONTENT_REPO_PATH, commit, metaData); err != nil {
			blog.Errorf("Error updating content: %v", err)
			return errors.New("error updating content: '" + metaData.ID + "', See server logs for more information")
		}
	}

	// update the layout with new meta data
	layout := layoutCache.Load().(Layout)
	metaDataMap := make(map[string]ContentMeta, len(newMetaDatas))
	for _, metaData := range newMetaDatas {
		metaDataMap[metaData.ID] = metaData
	}
	updateSidebarItems(layout.Sidebar, metaDataMap)
	updateFooterItems(layout.Footer, metaDataMap)
	SetLayout(&layout)

	// update the assets
	if err := updateAssets(commit); err != nil {
		blog.Errorf("Error updating assets: %v", err)
		return errors.New("error updating assets")
	}

	// run tailwind
	if err = RunTailwind(false); err != nil {
		blog.Errorf("Error running tailwindcss: %v", err)
		return errors.New("error running tailwindcss")
	}

	// cleanup the content
	if err := cleanupContent(newMetaDatas); err != nil {
		blog.Errorf("Error cleaning up content: %v", err)
		return errors.New("error cleaning up content")
	}

	return nil
}

func UpdateSandbox(newMD string) (string, error) {
	sandboxMutex.Lock()
	defer sandboxMutex.Unlock()
	if newMD == sandboxMD {
		return sandBoxHTML, nil
	}
	if newMD == "" {
		blog.Warn("UpdateSandbox called with empty markdown")
	}
	// update the sandbox markdown and convert to html
	sandboxMD = newMD
	var err error
	sandBoxHTML, err = utils.MdToHTML(sandboxMD)
	if err != nil {
		return "", fmt.Errorf("error converting markdown to html: %v", err)
	}
	// run tailwind
	if err = RunTailwind(true); err != nil {
		blog.Errorf("Error running tailwindcss: %v", err)
		return "", errors.New("error running tailwindcss")
	}
	return sandBoxHTML, nil
}

// GetMeta retrieves all content meta data.
func GetMeta() ([]ContentMeta, error) {
	var contentMeta []ContentMeta
	result := DB.Model(&ContentModel{}).Select("id", "commit", "rel_path").Find(&contentMeta)
	if result.Error != nil {
		return nil, result.Error
	}
	return contentMeta, nil
}

// RunTailwind runs the tailwindcss CLI to generate the CSS file.
func RunTailwind(sandboxOnly bool) error {
	tailwindMutex.Lock()
	defer tailwindMutex.Unlock()

	// get paths
	dbHtmlPath := filepath.Join(CONTENT_HTML_PATH, "db")
	sbHtmlPath := filepath.Join(CONTENT_HTML_PATH, "sandbox")

	// clean the sandbox directory, copy the html to it
	if err := files.CleanDir(sbHtmlPath); err != nil {
		blog.Errorf("Error cleaning the sandbox directory: %v", err)
		return errors.New("error cleaning the sandbox directory")
	}
	if err := files.CreateFile(filepath.Join(sbHtmlPath, "sandbox.html"), sandBoxHTML); err != nil {
		return err
	}

	// if not sandbox only, clean the db directory and copy all html content to it
	if !sandboxOnly {
		// clean the dir
		if err := files.CleanDir(dbHtmlPath); err != nil {
			blog.Errorf("Error cleaning the db directory: %v", err)
			return errors.New("error cleaning the db directory")
		}
		// copy all html content to it
		var contents []ContentModel
		result := DB.FindInBatches(&contents, 50, func(tx *gorm.DB, batch int) error {
			var i int64
			for i = 0; i < tx.RowsAffected; i++ {
				if err := files.CreateFile(filepath.Join(dbHtmlPath, fmt.Sprint(contents[i].ID)+".html"), contents[i].HTML); err != nil {
					return err
				}
			}
			return nil
		})
		if result.Error != nil {
			return result.Error
		}
	}

	// run the tailwindcss CLI
	tailInput := "./data/css/input.css"
	tailOutput := "./data/css/output.css"
	tailConfigPath := "./configs/tailwind.config.js"
	cmd := exec.Command("npx", "tailwindcss", "--config", tailConfigPath, "-i", tailInput, "-o", tailOutput, "--minify")
	if utils.DebugMode {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

// ==== Helper / Private Functions ============================================

func copyCommits(source, target []ContentMeta) {
	// key: id, value: commit
	commitMap := make(map[string]string, len(source))
	for _, item := range source {
		commitMap[item.ID] = item.Commit
	}
	// copy the commits
	for i, item := range target {
		if commit, exists := commitMap[item.ID]; exists {
			target[i].Commit = commit
		}
	}
}

// updateSidebarItems Recursively updates the meta data for all files in the sidebar.
func updateSidebarItems(items []SidebarItem, metaDataMap map[string]ContentMeta) {
	for i := range items {
		if items[i].Type == "file" {
			if meta, ok := metaDataMap[items[i].Meta.ID]; ok {
				items[i].Meta = meta
			} else {
				blog.Errorf("ID: '%s', not found for SidebarItem: %s", items[i].Meta.ID, items[i].Name)
				// TODO: webhook message
			}
		} else if items[i].Type == "folder" {
			updateSidebarItems(items[i].Contents, metaDataMap)
		}
	}
}

// updateFooterItems
func updateFooterItems(items []FooterItem, metaDataMap map[string]ContentMeta) {
	for i := range items {
		if items[i].Type == "file" {
			if meta, ok := metaDataMap[items[i].Meta.ID]; ok {
				items[i].Meta = meta
			} else {
				blog.Errorf("ID: '%s', not found for FooterItem: %s", items[i].Meta.ID, items[i].Name)
				// TODO: webhook message
			}
		}
	}
}

// extractIdFromFile extracts the ID from the first line of the file.
func extractIdFromFile(path string) (string, error) {
	line, err := files.ReadFirstLine(path)
	if err != nil {
		return "", err
	}
	parts := strings.Split(line, ":")
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "<!-- ID") || !strings.HasSuffix(parts[1], "-->") {
		return "", fmt.Errorf("invalid format")
	}
	return strings.TrimSpace(strings.TrimSuffix(parts[1], "-->")), nil
}

// loadIDs loads and parses the ids.json as well as set rel_path's to MISSING_FILE if the file does not exist or the ID does not match.
func loadIDs(contentPath string) ([]ContentMeta, error) {
	// read the ids file
	var fileContent map[string]string
	if exists, err := files.LoadJSON(filepath.Join(contentPath, ".github", "ids.json"), &fileContent); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("ids.json not found")
	}
	var metaDatas []ContentMeta
	for id, relPath := range fileContent {
		metaDatas = append(metaDatas, ContentMeta{ID: id, RelPath: relPath})
	}

	blog.Debugf("Loaded %d ids", len(metaDatas))

	// ensure rel_path are pointing to a file with the first line being `<!-- ID: TOKEN -->`, if not set to MISSING_FILE
	for i := range metaDatas {
		path := filepath.Join(contentPath, metaDatas[i].RelPath)
		// check if the file exists
		if exists, err := files.Exists(path); err != nil {
			return nil, err
		} else if !exists {
			metaDatas[i].RelPath = MISSING_FILE
			continue
		}
		fileID, err := extractIdFromFile(path)
		if err != nil {
			if err.Error() == "invalid format" {
				metaDatas[i].RelPath = MISSING_FILE
				continue
			} else {
				return nil, err
			}
		}
		if fileID != metaDatas[i].ID {
			metaDatas[i].RelPath = MISSING_FILE
		}
	}

	return metaDatas, nil
}

// updateContent updates the content for the given meta data.
func updateContent(repoPath, commit string, metaData ContentMeta) error {
	// handle missing pages
	if metaData.RelPath == MISSING_FILE {
		blog.Errorf("%s skipped, missing", metaData.ID)
		// TODO: webhook message
		return nil
	}
	// return if the file has not changed, else set the commit
	if changed, err := utils.GitFileDiff(repoPath, metaData.RelPath, metaData.Commit); err != nil {
		return err
	} else if !changed {
		blog.Debugf("%s skipped, no changes since %s", metaData.ID, metaData.Commit)
		return nil
	}
	metaData.Commit = commit
	// convert the md to html
	md, err := files.ReadFile(filepath.Join(repoPath, metaData.RelPath))
	if err != nil {
		return err
	}
	html, err := utils.MdToHTML(md)
	if err != nil {
		return err
	}
	// save the content
	time.Sleep(10 * time.Millisecond) // reduce db load
	err = DB.Save(&ContentModel{ContentMeta: metaData, HTML: html, MD: md}).Error
	if err == nil {
		blog.Debugf("%s updated", metaData.ID)
	}
	return err
}

// cleanupContent deletes all content records that are not in the given list of meta data.
func cleanupContent(metaDatas []ContentMeta) error {
	if len(metaDatas) == 0 {
		blog.Debugf("No ids provided, skipping cleanup")
		return nil
	}

	ids := make([]string, len(metaDatas))
	for i, metaData := range metaDatas {
		ids[i] = metaData.ID
	}

	result := DB.Where("id NOT IN ?", ids).Delete(&ContentModel{})
	if result.Error != nil {
		return result.Error
	}

	blog.Debugf("Deleted %d records", result.RowsAffected)
	return nil
}

// updateAssets updates the assets in the database and data/assets directory.
func updateAssets(commit string) error {
	AssetsMutex.Lock()
	defer AssetsMutex.Unlock()

	var assets []AssetModel
	if err := DB.Find(&assets).Error; err != nil {
		return err
	}

	// get all asset paths in the content repo
	var err error
	var contentRepoAssetPaths []string
	if contentRepoAssetPaths, err = files.ListAllFiles(filepath.Join(CONTENT_REPO_PATH, utils.Config.ContentRepo.AssetsDir)); err != nil {
		return err
	}
	for i, path := range contentRepoAssetPaths {
		if relPath, err := filepath.Rel(CONTENT_REPO_PATH, path); err != nil {
			return err
		} else {
			contentRepoAssetPaths[i] = relPath
		}
	}

	// remove assets from AssetModel slice and data/assets/ that no longer in the content repo
	for i := len(assets) - 1; i >= 0; i-- {
		if !utils.Contains(assets[i].ID, contentRepoAssetPaths) {
			target := filepath.Join("./data", "assets", assets[i].ID)
			if exists, err := files.Exists(target); err != nil {
				return err
			} else if exists {
				if err = os.Remove(filepath.Join("./data", "assets", assets[i].ID)); err != nil {
					blog.Errorf("Error removing asset: %v", err)
				}
			}
			assets = append(assets[:i], assets[i+1:]...)
		}
	}

	// add new assets to AssetModel slice, leave commit as empty string
	for _, path := range contentRepoAssetPaths {
		found := false
		for _, asset := range assets {
			if asset.ID == path {
				found = true
				break
			}
		}
		if !found {
			assets = append(assets, AssetModel{ID: path, Commit: ""})
		}
	}

	// for each if diff copy to data/assets/ and update commit
	for i := range assets {
		var err error
		var changed bool
		if changed, err = utils.GitFileDiff(CONTENT_REPO_PATH, assets[i].ID, assets[i].Commit); err != nil {
			return err
		}
		var exists bool
		dst := filepath.Join("./data", "assets", assets[i].ID)
		if exists, err = files.Exists(dst); err != nil {
			return err
		}
		if changed || !exists {
			assets[i].Commit = commit
			src := filepath.Join(CONTENT_REPO_PATH, assets[i].ID)
			if err := files.CopyFile(src, dst); err != nil {
				return err
			}
			blog.Debugf("Asset updated, src: %s, dst: %s", src, dst)
		} else {
			blog.Debugf("Asset %s skipped, no changes since %s", dst, assets[i].Commit)
		}
	}

	// update the db
	if err := DB.Exec("DELETE FROM asset_models").Error; err != nil {
		return err
	}
	if err := DB.Create(&assets).Error; err != nil {
		return err
	}

	return nil
}
