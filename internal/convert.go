package internal

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
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

// convertFrontMatter converts front matter from one format to another.
// frontMatter is the front matter string to convert.
// keyMap is a map of key mappings to apply during conversion.
// targetFormat is the target format for the converted front matter ("toml" or "yaml").
// The function returns the converted front matter string and any error that occurred.
func convertFrontMatter(frontMatter string, keyMap map[string]string, targetFormat string) (convertedFrontMatter string, err error) {
	var frontMatterMap map[string]interface{}
	err = yaml.Unmarshal([]byte(frontMatter), &frontMatterMap)
	if err != nil {
		return
	}
	convertedMap := make(map[string]interface{})
	for key, value := range frontMatterMap {
		convertedKey := keyMap[key]
		if convertedKey == "" {
			convertedKey = key
		}
		convertedMap[convertedKey] = value
	}
	switch targetFormat {
	case "yaml":
		bys, err := yaml.Marshal(convertedMap)
		if err != nil {
			return "", err
		}
		convertedFrontMatter = string(bys)
	case "toml":
		var buffer bytes.Buffer
		encoder := toml.NewEncoder(&buffer)
		err := encoder.Encode(convertedMap)
		if err != nil {
			return "", err
		}
		convertedFrontMatter = buffer.String()
	default:
		err = fmt.Errorf("invalid target format specified")
		return
	}
	return fmt.Sprintf("---\n%s---", convertedFrontMatter), nil
}

// convertMarkdown converts a Markdown file from one format to another.
// srcFile is the path of the source file to convert.
// dstDir is the path of the destination directory where the converted file will be saved.
// keyMap is a map of key mappings to apply during conversion.
// targetFormat is the target format for the converted front matter ("toml" or "yaml").
// errors is a channel where any errors that occur will be sent.
func convertMarkdown(srcFile, dstDir string, keyMap map[string]string, targetFormat string, errors chan<- error) {
	dstFile := filepath.Join(dstDir, filepath.Base(srcFile))

	content, err := os.ReadFile(srcFile)
	if err != nil {
		errors <- fmt.Errorf("error reading source file %s: %v", srcFile, err)
		return
	}

	splits := strings.SplitN(string(content), "---", 3)
	if len(splits) < 3 {
		errors <- fmt.Errorf("error parsing source file %s: not enough sections", srcFile)
		return
	}
	frontMatter := splits[1]
	body := splits[2]

	convertedFrontMatter, err := convertFrontMatter(frontMatter, keyMap, targetFormat)
	if err != nil {
		errors <- fmt.Errorf("error converting front matter in file %s: %v", srcFile, err)
		return
	}

	convertedContent := fmt.Sprintf("%s\n\n%s", convertedFrontMatter, body)
	err = os.WriteFile(dstFile, []byte(convertedContent), 0644)
	if err != nil {
		errors <- fmt.Errorf("error writing to destination file %s: %v", dstFile, err)
		return
	}

	// fmt.Printf("Converted %s to %s\n", srcFile, dstFile)
}

// ConvertPosts converts a directory of Markdown files from one format to another.
// srcDir is the path of the source directory containing the files to convert.
// dstDir is the path of the destination directory where the converted files will be saved.
// keyMap is a map of key mappings to apply during conversion.
// targetFormat is the target format for the converted front matter ("toml" or "yaml").
// The function returns any error that occurred.
func ConvertPosts(srcDir, dstDir string, keyMap map[string]string, targetFormat string) error {
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating destination directory %s: %v", dstDir, err)
	}

	var wg sync.WaitGroup
	errors := make(chan error)
	semaphore := make(chan struct{}, runtime.NumCPU())

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			wg.Add(1)
			semaphore <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-semaphore }()
				convertMarkdown(path, dstDir, keyMap, targetFormat, errors)
			}()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking source directory %s: %v", srcDir, err)
	}

	go func() {
		wg.Wait()
		close(errors)
	}()

	for err := range errors {
		fmt.Println(err)
	}

	return nil
}
