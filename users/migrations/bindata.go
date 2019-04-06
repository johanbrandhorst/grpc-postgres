package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var __1_initial_schema_down_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xf0\x74\x53\x70\x8d\xf0\x0c\x0e\x09\x56\x28\x2d\x4e\x2d\x2a\x56\x70\x76\x0c\x76\x76\x74\x71\xb5\xe6\x82\xa8\x89\x0c\x70\x55\x28\xca\xcf\x49\xb5\xe6\x02\x04\x00\x00\xff\xff\xa1\x96\x24\xe7\x34\x00\x00\x00")

func _1_initial_schema_down_sql() ([]byte, error) {
	return bindata_read(
		__1_initial_schema_down_sql,
		"1_initial_schema.down.sql",
	)
}

var __1_initial_schema_up_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\xce\xc1\x6a\xb4\x40\x0c\x07\xf0\xfb\x3c\xc5\xff\xa6\xc2\xbe\xc1\x77\x9a\x6f\x37\xa5\x43\x75\x94\x31\x43\x6b\x2f\x62\xd7\x20\x42\x47\x97\x51\x0f\x7d\xfb\xd2\xa1\x5d\x0a\xbd\x84\x84\xe4\x97\xe4\xec\x48\x33\x81\x5e\x98\x6c\x6b\x6a\x8b\xdb\x74\x8d\x1f\xb7\x7d\xfd\xa7\xd4\x77\x8f\xbb\x86\x10\xd7\x77\x81\x6e\x41\xd6\x57\xc8\xb3\xe9\x90\x6d\xcf\x4e\xc8\x82\x84\x37\x89\x5f\xd9\x30\x86\x79\xc9\x8a\x5f\x4e\xff\x2f\x09\xc7\x26\x71\x43\xae\x00\x60\x1e\xe1\xbd\xb9\xa0\x71\xa6\xd2\xae\xc3\x13\x75\xb8\xd0\x83\xf6\x25\x63\x92\xa5\x8f\xc3\x32\xae\xa1\x3f\x8e\x79\xcc\x8b\x53\x22\xe9\x6e\x0a\xb6\x66\x58\x5f\x96\x77\xf1\xf3\x44\x9a\xbb\x46\x19\x76\xe9\xf7\x39\x08\xd8\x54\xd4\xb2\xae\x1a\x3c\x1b\x7e\x4c\x25\x5e\x6b\x4b\x7f\x57\x9c\xbd\x73\x64\xb9\xbf\x0b\x55\xa8\xcf\x00\x00\x00\xff\xff\xba\xd1\x76\x78\x12\x01\x00\x00")

func _1_initial_schema_up_sql() ([]byte, error) {
	return bindata_read(
		__1_initial_schema_up_sql,
		"1_initial_schema.up.sql",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"1_initial_schema.down.sql": _1_initial_schema_down_sql,
	"1_initial_schema.up.sql": _1_initial_schema_up_sql,
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
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"1_initial_schema.down.sql": &_bintree_t{_1_initial_schema_down_sql, map[string]*_bintree_t{
	}},
	"1_initial_schema.up.sql": &_bintree_t{_1_initial_schema_up_sql, map[string]*_bintree_t{
	}},
}}
