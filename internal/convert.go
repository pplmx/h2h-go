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
		bys, err := yaml.Marshal(convertedMap)
		if err != nil {
			return "", err
		}
		convertedFrontMatter = string(bys)
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

func convertMarkdown(srcFile, dstDir string, keyMap map[string]string, targetFormat string, errors chan<- error) {
	dstFile := filepath.Join(dstDir, filepath.Base(srcFile))

	content, err := ioutil.ReadFile(srcFile)
	if err != nil {
		errors <- fmt.Errorf("error reading source file %s: %v", srcFile, err)
		return
	}

	splits := strings.Split(string(content), "---")
	frontMatter := splits[1]
	body := splits[2]

	convertedFrontMatter, err := convertFrontMatter(frontMatter, keyMap, targetFormat)
	if err != nil {
		errors <- fmt.Errorf("error converting front matter in file %s: %v", srcFile, err)
		return
	}

	convertedContent := fmt.Sprintf("%s\n\n%s", convertedFrontMatter, body)
	err = ioutil.WriteFile(dstFile, []byte(convertedContent), 0644)
	if err != nil {
		errors <- fmt.Errorf("error writing to destination file %s: %v", dstFile, err)
		return
	}

	fmt.Printf("Converted %s to %s\n", srcFile, dstFile)
}

func convertPosts(srcDir, dstDir string, keyMap map[string]string, targetFormat string) error {
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating destination directory %s: %v", dstDir, err)
	}

	var wg sync.WaitGroup
	errors := make(chan error)

	filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			wg.Add(1)
			go func() {
				defer wg.Done()
				convertMarkdown(path, dstDir, keyMap, targetFormat, errors)
			}()
		}
		return nil
	})

	go func() {
		wg.Wait()
		close(errors)
	}()

	for err := range errors {
		fmt.Println(err)
	}

	return nil
}
