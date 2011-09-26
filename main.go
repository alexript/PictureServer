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
        "crypto/sha1"
	"gocask"
)

// TODO: mutex for storage

type Error struct {
	Error os.Error
}

var (
        uploadTemplate, _ = template.ParseFile("templates/upload.html")
        editTemplate   *template.Template // set up in init()
        //postTemplate, _   = template.ParseFile("post.html")
        errorTemplate, _  = template.ParseFile("templates/error.html")

)

func main(){
	http.HandleFunc("/", errorHandler(upload))
	http.HandleFunc("/img", errorHandler(img))
	http.ListenAndServe(":8080",nil)
	
}

type Image struct {
        Data []byte
}

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


	const max = 600 // Масштабируем пропорционально до 600 пикселей, если какая-либо сторона больше. 
	if b := i.Bounds(); b.Dx() > max || b.Dy() > max {
		w, h := max, max
                if b.Dx() > b.Dy() {
			h = b.Dy() * h / b.Dx()
		} else {
			w = b.Dx() * w / b.Dy()
		}
	        i = resize.Resize(i, i.Bounds(), w, h)
        }

        // Encode as a new JPEG image.
	buf := new(bytes.Buffer)
	buf.Reset()
	err = jpeg.Encode(buf, i, nil)
	check(err)

	var barray []byte = buf.Bytes()
	var key string = keyOf(barray)
	storage, _ :=gocask.NewGocask("images/storage")
	err = storage.Put(key, barray)
	storage.Close()
	check(err)
	
	const maxth = 240 // Масштабируем пропорционально до 240 пикселей, если какая-либо сторона больше. 
	if b := i.Bounds(); b.Dx() > maxth || b.Dy() > maxth {
		w, h := maxth, maxth
                if b.Dx() > b.Dy() {
			h = b.Dy() * h / b.Dx()
		} else {
			w = b.Dx() * w / b.Dy()
		}
	        i = resize.Resize(i, i.Bounds(), w, h)
        }

        // Encode as a new JPEG image.
	buf2 := new(bytes.Buffer)
	buf2.Reset()
	err = jpeg.Encode(buf2, i, nil)
	check(err)

	var barrayb []byte = buf2.Bytes()
	var keyth string = "th-" + key
	storageth, _ :=gocask.NewGocask("images/storage")
	err = storageth.Put(keyth, barrayb)
	storageth.Close()
	check(err)
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	fmt.Fprintf(w,"{\"offerPicUrl\":\"img?id=" + key + "\",\"offerThumbUrl\":\"img?id=" + keyth + "\"}")
}

// keyOf returns (part of) the SHA-1 hash of the data, as a hex string.
func keyOf(data []byte) string {
        sha := sha1.New()
        sha.Write(data)
        return fmt.Sprintf("%x", string(sha.Sum())[0:8])
}

func img(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	storage, _ :=gocask.NewGocask("images/storage")
	buf, err := storage.Get(id)
	storage.Close()
	check(err)
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Expires", "-1")
	w.Write(buf)
}

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
