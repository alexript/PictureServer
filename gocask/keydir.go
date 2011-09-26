package gocask

import (
	"os"
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"fmt"
	"bufio"
)

const (
	/* Header offset for each record in the store.
	This offset contains the following information (in the giver order)
	| -------------------------------------------------------------------------|
	| crc (int32) | tstamp (int32) | key length (int32) | value length (int32) |
	| -------------------------------------------------------------------------|
	*/
	RECORD_HEADER_SIZE int32 = 16
)

/*
	Wrap a os.file and provide some convenient methods.
*/
type GFile struct {
	file *os.File
	cpos int32
}

/*
   Entries in the keydir, which holds the location of any key in the key-store
*/
type KeydirEntry struct {
	gfile   *GFile
	vsz    int32
	vpos   int32
	tstamp int64
}

/*
   In memory structure that holds the location of all the keys in the
   key-value store
*/
type Keydir struct {
	keys map[string]*KeydirEntry
}

/*
	Wrap the file f in an convenient strucutre.
*/
func NewGFile(f *os.File) *GFile {
	return &GFile{f, 0}
}

/*
	Instantiate an empty key dir
*/
func NewKeydir() *Keydir {
	ret := new(Keydir)
	ret.keys = make(map[string]*KeydirEntry)
	return ret
}

/*
	Store the information on the file, update the current pos and return the position and size of the value entry
*/
func (f *GFile) StoreData(key string, value []byte) (vpos int32, vsz int32, err os.Error) {
	buff := new(bytes.Buffer)
	keydata := []byte(key)
	binary.Write(buff, binary.BigEndian, int32(0x0aaaaaaa))
	binary.Write(buff, binary.BigEndian, int32(len(keydata)))
	binary.Write(buff, binary.BigEndian, int32(len(value)))
	buff.Write(keydata)
	buff.Write(value)

	crc := crc32.ChecksumIEEE(buff.Bytes())

	vpos = int32(RECORD_HEADER_SIZE /* crc + tstamp + len key data + len value */ + int32(len(keydata)))
	vsz = int32(len(value))
	buff2 := new(bytes.Buffer)
	binary.Write(buff2, binary.BigEndian, crc)
	buff2.Write(buff.Bytes())
	var sz int
	sz, err = f.file.Write(buff2.Bytes())
	vsz = int32(len(value))
	f.cpos += int32(sz)
	return vpos, vsz, err
}

/*
	Read the header structure from the file and return the header information.
	If data could not be obtained return an errro (including an os.EOF error)
*/
func (f *GFile) ReadHeader() (crc, tstamp, klen, vlen, vpos int32, key []byte, err os.Error) {
	var hdrbuff []byte = make([]byte, RECORD_HEADER_SIZE /* crc + tstamp + len key data + len value */ )
	var sz int
	sz, err = f.file.Read(hdrbuff)

	if err != nil {
		return
	}

	if int32(sz) != RECORD_HEADER_SIZE {
		err = os.NewError(fmt.Sprintf("Invalid header size. Expected %d got %d bytes", RECORD_HEADER_SIZE, sz))
	}

	buff := bufio.NewReader(bytes.NewBuffer(hdrbuff))
	binary.Read(buff, binary.BigEndian, &crc)
	binary.Read(buff, binary.BigEndian, &tstamp)
	binary.Read(buff, binary.BigEndian, &klen)
	binary.Read(buff, binary.BigEndian, &vlen)

	key = make([]byte, klen)
	sz, err = f.file.Read(key)

	if err != nil {
		return
	}

	if int32(sz) != klen {
		err = os.NewError(fmt.Sprintf("Invalid key size. Expected %d got %d bytes", klen, sz))
		return
	}

	f.file.Seek(int64(vlen), 1) /* move foward in the file to the next header (means skip the value) */

	vpos = f.cpos + RECORD_HEADER_SIZE + klen

	f.cpos += int32(RECORD_HEADER_SIZE + klen + vlen)

	return
}

/*
	Save the key/value pair in the given file f and update the keydir strucutre.
*/
func (kd *Keydir) WriteTo(f *GFile, key string, value []byte) os.Error {
	kde := new(KeydirEntry)
	var err os.Error

	if f == nil || f.file == nil {
		panic("file is nil")
	}

	kde.vpos, kde.vsz, err = f.StoreData(key, value)

	kde.gfile = f
	kde.tstamp = 0
	kd.keys[key] = kde
	return err
}

/*
	Populate the keydir structure with the information from the given file.
	Scan the entire file looking for information
*/
func (kd *Keydir) Fill(f *GFile) os.Error {
	f.file.Seek(0, 0) /* place the cursor in the begin of the file */
	var toreturn os.Error
	for {
		_, _, _, vsz, vpos, keydata, err := f.ReadHeader()

		if err != nil && err != os.EOF {
			toreturn = err
			break
		} else if err == os.EOF {
			break
		}

		key := string(keydata)
		kde := new(KeydirEntry)
		kde.vpos = vpos
		kde.vsz = vsz
		kde.tstamp = 0
		kde.gfile = f

		kd.keys[key] = kde

		if err == os.EOF {
			break
		}
	}

	return toreturn
}

func (kde *KeydirEntry) readValue() (value []byte, err os.Error) {
	value = make([]byte, kde.vsz)
	var read int
	read, err = kde.gfile.file.ReadAt(value, int64(kde.vpos))
	if int32(read) != kde.vsz {
		err = os.NewError(fmt.Sprintf("Expected %d bytes got %d", kde.vsz, read))
	}

	return
}
