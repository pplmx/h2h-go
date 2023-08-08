package internal

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var HEXO_TO_HUGO_KEY_MAP = map[string]string{
	"title":       "title",
	"date":        "date",
	"updated":     "lastmod",
	"categories":  "categories",
	"tags":        "tags",
	"description": "description",
	"keywords":    "keywords",
	"permalink":   "slug",
}

var HUGO_TO_HEXO_KEY_MAP = reverseMap(HEXO_TO_HUGO_KEY_MAP)

func reverseMap(m map[string]string) map[string]string {
	n := make(map[string]string, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

func convertFrontMatter(frontMatter string, keyMap map[string]string, targetFormat string) (string, error) {
	var frontMatterMap map[string]interface{}
	err := yaml.Unmarshal([]byte(frontMatter), &frontMatterMap)
	if err != nil {
		return "", err
	}
	convertedMap := make(map[string]interface{})
	for key, value := range frontMatterMap {
		convertedKey := keyMap[key]
		if convertedKey == "" {
			convertedKey = key
		}
		convertedMap[convertedKey] = value
	}
	var convertedFrontMatter string
	if targetFormat == "yaml" {
		bytes, err := yaml.Marshal(convertedMap)
		if err != nil {
			return "", err
		}
		convertedFrontMatter = string(bytes)
	} else if targetFormat == "toml" {
		var buffer bytes.Buffer
		encoder := toml.NewEncoder(&buffer)
		err := encoder.Encode(convertedMap)
		if err != nil {
			return "", err
		}
		convertedFrontMatter = buffer.String()
	} else {
		return "", fmt.Errorf("invalid target format specified")
	}
	return fmt.Sprintf("---\n%s---", convertedFrontMatter), nil
}

func convertMarkdown(srcFile, dstDir string, keyMap map[string]string, targetFormat string) {
	dstFile := filepath.Join(dstDir, filepath.Base(srcFile))

	content, err := ioutil.ReadFile(srcFile)
	if err != nil {
		fmt.Printf("Error reading source file %s: %v\n", srcFile, err)
		return
	}

	splits := strings.Split(string(content), "---")
	frontMatter := splits[1]
	body := splits[2]

	convertedFrontMatter, err := convertFrontMatter(frontMatter, keyMap, targetFormat)
	if err != nil {
		fmt.Printf("Error converting front matter in file %s: %v\n", srcFile, err)
		return
	}

	convertedContent := fmt.Sprintf("%s\n\n%s", convertedFrontMatter, body)
	err = ioutil.WriteFile(dstFile, []byte(convertedContent), 0644)
	if err != nil {
		fmt.Printf("Error writing to destination file %s: %v\n", dstFile, err)
		return
	}

	fmt.Printf("Converted %s to %s\n", srcFile, dstFile)
}

func convertPosts(srcDir, dstDir string, keyMap map[string]string, targetFormat string) {
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		fmt.Printf("Error creating destination directory %s: %v\n", dstDir, err)
		return
	}
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		fmt.Printf("Error reading source directory %s: %v\n", srcDir, err)
		return
	}

	var wg sync.WaitGroup

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".md") {
			srcFile := filepath.Join(srcDir, f.Name())
			wg.Add(1)
			go func(srcFile string) {
				defer wg.Done()
				convertMarkdown(srcFile, dstDir, keyMap, targetFormat)
			}(srcFile)
		}
	}

	wg.Wait()
}
