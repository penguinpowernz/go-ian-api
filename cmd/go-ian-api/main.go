package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/upload", uploadFileHandler())

	fs := http.FileServer(http.Dir(""))
	http.Handle("/files/", http.StripPrefix("/files", fs))

	log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading files.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// TODO: rate limit

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20) // limit your max input length!
		// in your case file would be fileupload
		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		if header.Size > 50*1024*1024 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		name := strings.Split(header.Filename, ".")
		fmt.Printf("File name %s\n", name[0])
		ext := strings.Join(name[1:], ".")

		switch ext {
		case "tar.gz", "zip":
			break
		default:
			// TODO: error out
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// TODO: ESCAPE filename
		savedFilePath := "/pathToStoreFile/" + header.Filename
		f, err := os.OpenFile(savedFilePath, os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Copy the file to the destination path
		io.Copy(f, file)

		// TODO: extract file contents

		// TODO: run ian on the file contents

		// TODO: move debian package to unique location

		// TODO: redirect to debian package location

		return
	})
}
