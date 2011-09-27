package main

import (
	"bytes"
        "fmt"
        "http"
	"image"
	"image/jpeg"
	_ "image/png" // import so we can read PNG files.
        "os"
        "template"
	"resize"
        "crypto/md5"
	"gocask"
	"time"
	"io"
)

// TODO: mutex for storage

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

)

// Surprise!!!
func main(){
	http.HandleFunc("/", errorHandler(upload))
	http.HandleFunc("/img", errorHandler(img))
	http.HandleFunc("/tmb", errorHandler(tmb))
	http.ListenAndServe(":8080",nil)
	
}

// Proportional image resize to max size by any side
func resizeImage(i image.Image, max int) (image.Image) {
	if b := i.Bounds(); b.Dx() > max || b.Dy() > max {
		w, h := max, max
                if b.Dx() > b.Dy() {
			h = b.Dy() * h / b.Dx()
		} else {
			w = b.Dx() * w / b.Dy()
		}
	        i = resize.Resize(i, i.Bounds(), w, h)
        }
	return i
}

// resize image, store and return resized image
func storeImage(key string, i image.Image, maxsize int, storename string) (image.Image) {
	// store big image
	i = resizeImage(i, maxsize) // Масштабируем пропорционально до maxsize пикселей, если какая-либо сторона больше. 

        // Encode as a new JPEG image.
	buf := new(bytes.Buffer)
	buf.Reset()
	err := jpeg.Encode(buf, i, nil)
	check(err)

	var barray []byte = buf.Bytes()
	storage, _ :=gocask.NewGocask("images/" + storename)
	err = storage.Put(key, barray)
	storage.Close()
	check(err)
	return i
}

// Catch post request, decode image and store it
func upload(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
                // No upload; show the upload form.
                uploadTemplate.Execute(w, nil)
                return
        }

        f, _, err := r.FormFile("image")
        check(err)
        defer f.Close()

        // Grab the image data
	i, _, err := image.Decode(f)
        check(err)

	var key string = keyOf()

	// store image
	i = storeImage(key, i, 600, "storage")
	i = storeImage(key, i, 240, "thumbs")

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

// read image by key from storage
func readImage(key string, storagename string) ([]byte) {
	storage, _ :=gocask.NewGocask("images/" + storagename)
	buf, err := storage.Get(key)
	storage.Close()
	check(err)
	return buf
}

// catch /img request and return image
func img(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	buf := readImage(id, "storage")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Write(buf)
}

// catch /tmb request and return thumbnail
func tmb(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	buf := readImage(id, "thumbs")
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

// check aborts the current execution if err is non-nil.
func check(err os.Error) {
        if err != nil {
                panic(err)
        }
}
