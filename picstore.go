package picstore
// image storage packet

import (
	"image"
	"resize"
	"bytes"
	"image/jpeg"
	"gocask"
	"errors"
	"sync"
)

var (
	mu sync.Mutex
)

func cropImage(i image.Image, vert int, hor int) (image.Image){
	b := i.Bounds()
	
	topy := (b.Dy() - vert)/2
	boty := b.Dy() - topy
	leftx := (b.Dx() - hor)/2
	rightx := b.Dx() - leftx
	
	rect := image.Rectangle{image.Point{leftx, topy}, image.Point{rightx, boty}}
	irgba := i.(*image.RGBA)
	img := irgba.SubImage(rect).(*image.RGBA)
	
	return img
	
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

func StoreSubImage(key string, i image.Image, vert int, hor int, storename string){
	
	img := cropImage(i, vert, hor)
	
	buf := new(bytes.Buffer)
	buf.Reset()
	err := jpeg.Encode(buf, img, nil)
	errors.Check(err)

	var barray []byte = buf.Bytes()
	mu.Lock()
	storage, _ :=gocask.NewGocask("images/" + storename)
	err = storage.Put(key, barray)
	storage.Close()
	mu.Unlock()
	errors.Check(err)


}

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
	mu.Lock()
	storage, _ :=gocask.NewGocask("images/" + storename)
	err = storage.Put(key, barray)
	storage.Close()
	mu.Unlock()
	errors.Check(err)
	return i
}

// read image by key from storage
func Read(key string, storagename string) ([]byte) {
	mu.Lock()
	storage, _ :=gocask.NewGocask("images/" + storagename)
	buf, err := storage.Get(key)
	storage.Close()
	mu.Unlock()
	errors.Check(err)
	return buf
}

