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
	c.SaveUploadedFile(file, path.Join(svr.UploadPath, header.Filename))

	// TODO: extract file contents

	// TODO: run ian on the file contents

	// TODO: move debian package to unique location

	// TODO: redirect to debian package location
}
