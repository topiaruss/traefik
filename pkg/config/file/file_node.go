package file

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/pkg/config/parser"
	"gopkg.in/yaml.v2"
)

// decodeFileToNode decodes the configuration in filePath in a tree of untyped nodes.
// If filters is not empty, it skips any configuration element whose name is
// not among filters.
func decodeFileToNode(filePath string, filters ...string) (*parser.Node, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})

	switch filepath.Ext(filePath) {
	case ".toml":
		err = toml.Unmarshal(content, &data)
		if err != nil {
			return nil, err
		}

	case ".yml", ".yaml":
		var err error
		err = yaml.Unmarshal(content, data)
		if err != nil {
			return nil, err
		}

		return decodeRawToNode(data, filters...)

	default:
		return nil, fmt.Errorf("unsupported file extension: %s", filePath)
	}

	return decodeRawToNode(data, filters...)
}

func getRootFieldNames(element interface{}) []string {
	if element == nil {
		return nil
	}

	rootType := reflect.TypeOf(element)

	return getFieldNames(rootType)
}

func getFieldNames(rootType reflect.Type) []string {
	var names []string

	if rootType.Kind() == reflect.Ptr {
		rootType = rootType.Elem()
	}

	if rootType.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < rootType.NumField(); i++ {
		field := rootType.Field(i)

		if !parser.IsExported(field) {
			continue
		}

		if field.Anonymous &&
			(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct || field.Type.Kind() == reflect.Struct) {
			names = append(names, getFieldNames(field.Type)...)
			continue
		}

		names = append(names, field.Name)
	}

	return names
}
