package main

import (
        "fmt"
        "http"
	"image"
	_ "image/jpeg"
	_ "image/png" // import so we can read PNG files.
        "os"
        "template"
        "crypto/md5"
	"time"
	"io"
	"picstore"
	"errors"
	"sync"
)

// Error container struct for error template.
type Error struct {
	Error os.Error
}

// global vars
var (
        uploadTemplate, _ = template.ParseFile("templates/upload.html")
        editTemplate   *template.Template // set up in init()
        //postTemplate, _   = template.ParseFile("post.html")
        errorTemplate, _  = template.ParseFile("templates/error.html")
	mu sync.Mutex
)

// Surprise!!!
func main(){
	http.HandleFunc("/", errorHandler(upload))
	http.HandleFunc("/img", errorHandler(img))
	http.HandleFunc("/tmb", errorHandler(tmb))
	http.ListenAndServe(":8080",nil)
	
}



// Catch post request, decode image and store it
func upload(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
                // No upload; show the upload form.
                uploadTemplate.Execute(w, nil)
                return
        }

        f, _, err := r.FormFile("image")
        errors.Check(err)
        defer f.Close()

        // Grab the image data
	i, _, err := image.Decode(f)
        errors.Check(err)

	var key string = keyOf()

	// store image
	mu.Lock()
	i = picstore.Store(key, i, 600, "storage")
	i = picstore.Store(key, i, 240, "thumbs")
	mu.Unlock()
	// generate result

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	fmt.Fprintf(w,"{\"offerPicUrl\":\"img?id=" + key + "\",\"offerThumbUrl\":\"tmb?id=" + key + "\"}")
}

// keyOf returns the MD5 hash of the current time in nanoseconds, as a hex string.
func keyOf() string {
        md := md5.New()
	io.WriteString(md, fmt.Sprintf("%u", time.Nanoseconds()))
	return fmt.Sprintf("%x", string(md.Sum()))
}


// catch /img request and return image
func img(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	mu.Lock()
	buf := picstore.Read(id, "storage")
	mu.Unlock()
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Write(buf)
}

// catch /tmb request and return thumbnail
func tmb(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	mu.Lock()
	buf := picstore.Read(id, "thumbs")
	mu.Unlock()
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Write(buf)
}

// error handler. Display error template with error message.
func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                defer func() {
                        if err, ok := recover().(os.Error); ok {
                                w.WriteHeader(http.StatusInternalServerError)
				error := &Error{Error: err}
                                errorTemplate.Execute(w, error)
                        }
                }()
                fn(w, r)
        }
}

