// Code generated by go-bindata. DO NOT EDIT.
//  memcopy: true
//  compress: true
//  decompress: once
//  asset-dir: true
//  restore: true
// sources:
//  users/migrations/001_initial_schema.down.sql
//  users/migrations/001_initial_schema.up.sql

package migrations

import (
	"bytes"
	"compress/flate"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tmthrgd/go-bindata/restore"
)

type asset struct {
	name string
	data string
	size int64

	once  sync.Once
	bytes []byte
	err   error
}

func (a *asset) Name() string {
	return a.name
}

func (a *asset) Size() int64 {
	return a.size
}

func (a *asset) Mode() os.FileMode {
	return 0
}

func (a *asset) ModTime() time.Time {
	return time.Time{}
}

func (*asset) IsDir() bool {
	return false
}

func (*asset) Sys() interface{} {
	return nil
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]*asset{
	"001_initial_schema.down.sql": &asset{
		name: "001_initial_schema.down.sql",
		data: "" +
			"\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xf0\x74\x53\x70\x8d\xf0\x0c\x0e\x09\x56\x28\x2d\x4e" +
			"\x2d\x2a\x56\x70\x76\x0c\x76\x76\x74\x71\xb5\xe6\x82\xa8\x89\x0c\x70\x55\x28\xca\xcf\x49\xb5\xe6" +
			"\x02\x04\x00\x00\xff\xff",
		size: 52,
	},
	"001_initial_schema.up.sql": &asset{
		name: "001_initial_schema.up.sql",
		data: "" +
			"\x64\xce\xc1\x6a\xc3\x30\x0c\x06\xe0\xbb\x9f\xe2\xbf\x25\x81\xbe\x41\x4f\x5e\xab\xb1\xb0\xc4\x0d" +
			"\xae\xcc\x96\x5d\x42\xb6\x88\x12\x98\x5d\x70\xe2\xf7\x1f\xf5\x48\x2e\xbd\x08\x09\xf4\xfd\xd2\xc9" +
			"\x92\x66\x02\xf7\x1d\x21\xde\x7f\x05\xfa\x0a\x32\xae\x45\x59\xdc\x92\x2c\x6b\x71\x40\xe1\xc5\x7f" +
			"\x4b\x7c\x74\xe3\xe4\xe7\x50\x54\x47\xa5\x36\xa7\x5f\x1a\x42\x5a\x24\x2e\x28\x15\x00\xcc\x13\x9c" +
			"\xab\xcf\xe8\x6c\xdd\x6a\xdb\xe3\x9d\x7a\x9c\xe9\x55\xbb\x86\x71\x93\x30\xc4\x31\x4c\x77\x3f\xa4" +
			"\x34\x4f\x65\x75\xc8\x24\xdf\xcd\xc5\x5c\x18\xc6\x35\xcd\x2e\xb6\x27\xf2\xde\x4f\x94\x71\x95\x61" +
			"\x9d\xbd\x80\xeb\x96\xae\xac\xdb\x0e\x1f\x35\xbf\xe5\x11\x5f\x17\x43\xcf\x11\x27\x67\x2d\x19\x1e" +
			"\x76\xf1\x1f\x16\xc6\x47\x0a\x7d\xf2\x2e\x54\x75\x54\x7f\x01\x00\x00\xff\xff",
		size: 271,
	},
}

// AssetAndInfo loads and returns the asset and asset info for the
// given name. It returns an error if the asset could not be found
// or could not be loaded.
func AssetAndInfo(name string) ([]byte, os.FileInfo, error) {
	a, ok := _bindata[filepath.ToSlash(name)]
	if !ok {
		return nil, nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	a.once.Do(func() {
		fr := flate.NewReader(strings.NewReader(a.data))

		var buf bytes.Buffer
		if _, a.err = io.Copy(&buf, fr); a.err != nil {
			return
		}

		if a.err = fr.Close(); a.err == nil {
			a.bytes = buf.Bytes()
		}
	})
	if a.err != nil {
		return nil, nil, &os.PathError{Op: "read", Path: name, Err: a.err}
	}

	return a.bytes, a, nil
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	a, ok := _bindata[filepath.ToSlash(name)]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	data, _, err := AssetAndInfo(name)
	return data, err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}

	return names
}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	return restore.Asset(dir, name, AssetAndInfo)
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	return restore.Assets(dir, name, AssetDir, AssetAndInfo)
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree

	if name != "" {
		var ok bool
		for _, p := range strings.Split(filepath.ToSlash(name), "/") {
			if node, ok = node[p]; !ok {
				return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
			}
		}
	}

	if len(node) == 0 {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	rv := make([]string, 0, len(node))
	for name := range node {
		rv = append(rv, name)
	}

	return rv, nil
}

type bintree map[string]bintree

var _bintree = bintree{
	"001_initial_schema.down.sql": bintree{},
	"001_initial_schema.up.sql":   bintree{},
}
