package picstore
// image storage packet

import (
	"image"
	"resize"
	"bytes"
	"image/jpeg"
	"gocask"
	"errors"
)

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

// TODO: mutex for storage



// resize image, store and return resized image
func Store(key string, i image.Image, maxsize int, storename string) (image.Image) {
	// store big image
	i = resizeImage(i, maxsize) // Масштабируем пропорционально до maxsize пикселей, если какая-либо сторона больше. 

        // Encode as a new JPEG image.
	buf := new(bytes.Buffer)
	buf.Reset()
	err := jpeg.Encode(buf, i, nil)
	errors.Check(err)

	var barray []byte = buf.Bytes()
	storage, _ :=gocask.NewGocask("images/" + storename)
	err = storage.Put(key, barray)
	errors.Check(err)
	defer storage.Close()
	return i
}

// read image by key from storage
func Read(key string, storagename string) ([]byte) {
	storage, _ :=gocask.NewGocask("images/" + storagename)
	buf, err := storage.Get(key)
	errors.Check(err)
	defer storage.Close()
	return buf
}

