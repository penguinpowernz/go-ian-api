package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/satori/uuid"

	"github.com/gin-gonic/gin"
)

func main() {
	var uploadPath, extractPath, downloadPath string
	var maxFileSize int64

	flag.StringVar(&uploadPath, "u", "uploads", "Directory to upload files to")
	flag.StringVar(&extractPath, "e", "extracts", "Directory to extract into")
	flag.StringVar(&downloadPath, "d", "downloads", "Directory to serve downloads from")
	flag.Int64Var(&maxFileSize, "m", 50, "Max file size to serve (mB)")
	flag.Parse()

	checkFile(uploadPath)
	log.Println("found uploads path at", uploadPath)
	checkFile(extractPath)
	log.Println("found extracts path at", extractPath)
	checkFile(downloadPath)
	log.Println("found downloads path at", downloadPath)

	api := gin.Default()
	api.MaxMultipartMemory = maxFileSize << 20

	svr := &server{
		uploadPath,
		extractPath,
		downloadPath,
	}
	svr.AttachRoutes(api)
	// TODO: rate limit

	log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading files.")
	log.Fatal(api.Run(":8080"))
}

type server struct {
	UploadPath   string
	ExtractPath  string
	DownloadPath string
}

func (svr *server) AttachRoutes(r gin.IRouter) {
	r.POST("/upload", svr.UploadFile)
	r.GET("/download/:id/:filename", svr.DownloadFile)
}

func (svr *server) DownloadFile(c *gin.Context) {
	id := c.Param("id")
	fn := c.Param("filename")
	// TODO: escape the file path
	fp := path.Join(svr.DownloadPath, id, fn)

	c.FileAttachment(fp, fn)
}

func (svr *server) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	name := strings.Split(file.Filename, ".")
	ext := strings.Join(name[1:], ".")

	switch ext {
	case "tar.gz", "zip":
		break
	default:
		c.AbortWithError(400, errors.New("bad file type"))
		return
	}

	// TODO: escape file name if gin does not
	filepath := path.Join(svr.UploadPath, file.Filename)
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.AbortWithError(500, err)
		return
	}

	uid := unique()
	exdir := path.Join(svr.ExtractPath, uid)
	if err := os.MkdirAll(exdir, 0755); err != nil {
		c.AbortWithError(500, err)
		return
	}

	// extract file contents
	switch ext {
	case "tar.gz":
		err = extractTar(filepath, exdir)
	case "zip":
		err = extractZip(filepath, exdir)
	}

	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	if err := validateDir(exdir); err != nil {
		c.AbortWithError(400, err)
		return
	}

	// run ian on the file contents
	debpath, err := buildDpkg(exdir)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	log.Println("built debian package to", debpath)

	// move debian package to unique location
	dldir := path.Join(svr.DownloadPath, uid)
	if err := os.MkdirAll(dldir, 0755); err != nil {
		c.AbortWithError(500, err)
		return
	}
	log.Println("created download dir", dldir)

	dlpath := path.Join(dldir, path.Base(debpath))
	dlfile, err := os.OpenFile(dlpath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	defer dlfile.Close()
	log.Println("opened download path", dlpath)

	debfile, err := os.Open(debpath)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	defer debfile.Close()
	log.Println("opened debfile path", debpath)

	if _, err := io.Copy(dlfile, debfile); err != nil {
		c.AbortWithError(500, err)
		return
	}

	// redirect to debian package location

	c.Redirect(http.StatusSeeOther, path.Join("/download", uid, path.Base(debpath)))
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

func buildDpkg(dir string) (string, error) {
	cmd := exec.Command("ian", "pkg")
	cmd.Env = append(cmd.Env, "IAN_DIR="+dir)
	fn, err := cmd.Output()
	return strings.TrimSpace(string(fn)), err
}

func unique() string {
	u := uuid.NewV4()
	return u.String()
}

func checkFile(d string) {
	s, err := os.Stat(d)
	if !s.IsDir() {
		panic(d + " is not a dir")
	}

	if err != nil {
		panic(err)
	}
}
