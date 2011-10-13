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
)

// Error container struct for error template.
type Error struct {
	Error os.Error
}

// global vars
var (
        uploadTemplate, _ = template.ParseFile("templates/upload.html")
		uploadUserTemplate, _ = template.ParseFile("templates/uploaduser.html")
        errorTemplate, _  = template.ParseFile("templates/error.html")
)

// Surprise!!!
func main(){
	http.HandleFunc("/", errorHandler(upload))
	http.HandleFunc("/usr", errorHandler(uploadUser))
	http.HandleFunc("/img", errorHandler(img))
	http.HandleFunc("/tmb", errorHandler(tmb))
	http.HandleFunc("/uimg", errorHandler(uimg))
	http.HandleFunc("/utmb", errorHandler(utmb))
	http.HandleFunc("/crp", errorHandler(crp))
	http.ListenAndServe(":80",nil)
	
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
	i = picstore.Store(key, i, 640, "storage")
	picstore.StoreSubImage(key, i, 184, 640, "crop")
	i = picstore.Store(key, i, 210, "thumbs")
	// generate result

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	fmt.Fprintf(w,"{\"offerPicUrl\":\"img?id=" + key + "\",\"offerThumbUrl\":\"tmb?id=" + key + "\",\"offerLineThumbUrl\":\"crp?id=" + key + "\"}")
}

func uploadUser(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
                // No upload; show the upload form.
                uploadUserTemplate.Execute(w, nil)
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
	i = picstore.Store(key, i, 320, "userimg")
	i = picstore.Store(key, i, 180, "userthumb")
	// generate result

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	fmt.Fprintf(w,"{\"profilePicUrl\":\"uimg?id=" + key + "\",\"profileThumbUrl\":\"utmb?id=" + key + "\"}")
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
	buf := picstore.Read(id, "storage")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Write(buf)
}

func uimg(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	buf := picstore.Read(id, "userimg")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Write(buf)
}

func utmb(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	buf := picstore.Read(id, "userthumb")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Write(buf)
}

func crp(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	buf := picstore.Read(id, "crop")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Write(buf)
}

// catch /tmb request and return thumbnail
func tmb(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	buf := picstore.Read(id, "thumbs")
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

