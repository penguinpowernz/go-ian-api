package main

import (
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	MaxFileSize = 50 * 1024 * 1024
	UploadPath  = "uploads"
	ExtractPath = "extracts"
)

func main() {
	api := gin.Default()

	api.StaticFS("/download", http.Dir("downloads"))
	api.MaxMultipartMemory = 8 << 20

	svr := &server{}
	svr.AttachRoutes(api)
	// TODO: rate limit

	log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading files.")
	log.Fatal(api.Run(":8080"))
}

type server struct {
	UploadPath  string
	ExtractPath string
}

func (svr *server) AttachRoutes(r gin.IRouter) {
	r.POST("/upload", svr.UploadFile)
}

func (svr *server) UploadFile(c *gin.Context) {
	file, header := c.FormFile("file")
	if header.Size > MaxFileSize {
		c.AbortWithStatus(400)
		return
	}

	name := strings.Split(header.Filename, ".")
	ext := strings.Join(name[1:], ".")

	switch ext {
	case "tar.gz", "zip":
		break
	default:
		// TODO: error out
		c.AbortWithStatus(400)
		return
	}

	// TODO: escape file name if gin does not
	filepath := path.Join(svr.UploadPath, header.Filename)
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.AbortWithStatus(500)
		return
	}

	uid := unique()
	exdir := path.Join(svr.ExtractPath, uid)
	if err := os.MkdirAll(exdir, 0755); err != nil {
		c.AbortWithStatus(500)
		return
	}

	// extract file contents
	switch ext {
	case "tar.gz":
		err = extractTar(filepath, exdir)
	case "zip":
		err = extractZip(filepath, exdir)
	}

	if err := validateDir(exdir); err != nil {
		c.AbortWithStatus(400)
		return
	}

	// TODO: redirect to debian package location
}

func extractZip(filepath, dir string) error {
	cmd := exec.Command("unzip", "-d", dir, filepath)
	return cmd.Run()
}

func extractTar(filepath, dir string) error {
	cmd := exec.Command("tar", "-xzf", filepath, "-C", dir)
	return cmd.Run()
}

func validateDir(dir string) error {
	_, err := os.Stat(path.Join(dir, "DEBIAN", "control"))
	return err
}

func unique() string {
	u, _ := uuid.NewV4()
	return u.String()
}
