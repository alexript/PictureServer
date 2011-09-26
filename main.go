package main

import (
	"bytes"
        "fmt"
        "http"
	"image"
	"image/jpeg"
	_ "image/png" // import so we can read PNG files.
        "io"
	"io/ioutil"
        "os"
        "template"
	"resize"
        "crypto/sha1"
)

type Error struct {
	Error os.Error
}

var (
        uploadTemplate = template.MustParseFile("templates/upload.html", nil)
        editTemplate   *template.Template // set up in init()
        //postTemplate   = template.MustParseFile("post.html", nil)
        errorTemplate  = template.MustParseFile("templates/error.html", nil)
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
	// println(imageformatname)
        check(err)


	const max = 300 // Масштабируем пропорционально до 300 пикселей, если какая-либо сторона больше. 
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
	err = jpeg.Encode(buf, i, nil)
	check(err)

	resizedfile, err := ioutil.TempFile("./images", "ltl-")
	check(err)
	defer resizedfile.Close()
	io.Copy(resizedfile, buf)


	http.Redirect(w, r, "/img?id="+resizedfile.Name(), 302)
}

// keyOf returns (part of) the SHA-1 hash of the data, as a hex string.
func keyOf(data []byte) string {
        sha := sha1.New()
        sha.Write(data)
        return fmt.Sprintf("%x", string(sha.Sum())[0:8])
}

func img(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeFile(w, r, r.FormValue("id"))
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
