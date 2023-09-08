package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type File struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime string
	IsDir   bool
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Path[1:]
		filePath = filepath.Join("./db/", filePath)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		files, err := ioutil.ReadDir(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var fileList []File
		for _, file := range files {
			fileList = append(fileList, File{
				Name:    file.Name(),
				Size:    file.Size(),
				ModTime: file.ModTime().Format("02 Jan 2006 15:04:05"),
				IsDir:   file.IsDir(),
			})
		}

		tmpl := template.Must(template.ParseFiles("index.html"))
		tmpl.Execute(w, fileList)
	})
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/rename", renameHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/mkdir", mkdirHandler)
	http.ListenAndServe(":8080", nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileExt := filepath.Ext(fileHeader.Filename)
	originalFileName := strings.TrimSuffix(filepath.Base(fileHeader.Filename), filepath.Ext(fileHeader.Filename))
	filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + fileExt
	dst, err := os.Create("./db/" + filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "uploaded\n")

}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "file parameter is missing", http.StatusBadRequest)
		return
	}

	err := os.Remove(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "file deleted successfully")

}
func renameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	oldFilePath := r.URL.Query().Get("old")
	newFilePath := r.URL.Query().Get("new")

	if oldFilePath == "" || newFilePath == "" {
		http.Error(w, "old and new file parameters are required", http.StatusBadRequest)
		return
	}

	err := os.Rename(oldFilePath, newFilePath+filepath.Ext(oldFilePath))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to rename file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "file renamed successfully")

}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "file path is required", http.StatusBadRequest)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, fileName := filepath.Split(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func mkdirHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	dirName := r.FormValue("dirName")

	err := os.Mkdir(dirName, 0755)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Directory created successfully")
}
