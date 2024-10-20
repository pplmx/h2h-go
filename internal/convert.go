package internal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// KeyMap defines the mapping between different front matter formats
type KeyMap map[string]string

var (
	HexoToHugoKeyMap = KeyMap{
		"title":       "title",
		"date":        "date",
		"updated":     "lastmod",
		"categories":  "categories",
		"tags":        "tags",
		"description": "description",
		"keywords":    "keywords",
		"permalink":   "slug",
	}
	HugoToHexoKeyMap = reverseMap(HexoToHugoKeyMap)
)

func reverseMap(m KeyMap) KeyMap {
	n := make(KeyMap, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

// FrontMatterConverter handles the conversion of front matter
type FrontMatterConverter struct {
	keyMap       KeyMap
	targetFormat string
}

// NewFrontMatterConverter creates a new FrontMatterConverter
func NewFrontMatterConverter(keyMap KeyMap, targetFormat string) *FrontMatterConverter {
	return &FrontMatterConverter{
		keyMap:       keyMap,
		targetFormat: targetFormat,
	}
}

func (fmc *FrontMatterConverter) Convert(frontMatter string) (string, error) {
	var frontMatterMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontMatter), &frontMatterMap); err != nil {
		return "", fmt.Errorf("error unmarshaling front matter: %w", err)
	}

	convertedMap := make(map[string]interface{}, len(frontMatterMap))
	for key, value := range frontMatterMap {
		if convertedKey := fmc.keyMap[key]; convertedKey != "" {
			convertedMap[convertedKey] = value
		} else {
			convertedMap[key] = value
		}
	}

	var buf bytes.Buffer
	var err error

	switch fmc.targetFormat {
	case "yaml":
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(4)
		err = encoder.Encode(convertedMap)
	case "toml":
		err = toml.NewEncoder(&buf).Encode(convertedMap)
	default:
		return "", fmt.Errorf("invalid target format specified: %s", fmc.targetFormat)
	}

	if err != nil {
		return "", fmt.Errorf("error encoding front matter: %w", err)
	}

	return fmt.Sprintf("---\n%s---", buf.String()), nil
}

// MarkdownConverter handles the conversion of markdown files
type MarkdownConverter struct {
	fmc *FrontMatterConverter
}

// NewMarkdownConverter creates a new MarkdownConverter
func NewMarkdownConverter(keyMap KeyMap, targetFormat string) *MarkdownConverter {
	return &MarkdownConverter{
		fmc: NewFrontMatterConverter(keyMap, targetFormat),
	}
}

func (mc *MarkdownConverter) Convert(r io.Reader, w io.Writer) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading content: %w", err)
	}

	splits := strings.SplitN(string(content), "---", 3)
	if len(splits) < 3 {
		return fmt.Errorf("error parsing content: not enough sections")
	}

	convertedFrontMatter, err := mc.fmc.Convert(splits[1])
	if err != nil {
		return fmt.Errorf("error converting front matter: %w", err)
	}

	_, err = fmt.Fprintf(w, "%s\n\n%s", convertedFrontMatter, splits[2])
	return err
}

// ConversionError represents an error that occurred during the conversion process
type ConversionError struct {
	SourceFile string
	Err        error
}

func (e *ConversionError) Error() string {
	return fmt.Sprintf("error converting file %s: %v", e.SourceFile, e.Err)
}

// ConvertPosts converts all markdown posts in the source directory to the target format
func ConvertPosts(srcDir, dstDir string, keyMap KeyMap, targetFormat string) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("error creating destination directory %s: %w", dstDir, err)
	}

	mc := NewMarkdownConverter(keyMap, targetFormat)

	var wg sync.WaitGroup
	errChan := make(chan *ConversionError, runtime.NumCPU())
	semaphore := make(chan struct{}, runtime.NumCPU())

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			wg.Add(1)
			semaphore <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-semaphore }()

				if err := convertFile(mc, path, dstDir); err != nil {
					errChan <- &ConversionError{SourceFile: path, Err: err}
				}
			}()
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking source directory %s: %w", srcDir, err)
	}

	wg.Wait()
	close(errChan)

	var conversionErrors []*ConversionError
	for err := range errChan {
		fmt.Printf("Error converting file %s: %v\n", err.SourceFile, err.Err)
		conversionErrors = append(conversionErrors, err)
	}

	if len(conversionErrors) > 0 {
		return fmt.Errorf("encountered %d errors during conversion", len(conversionErrors))
	}

	return nil
}

func convertFile(mc *MarkdownConverter, srcPath, dstDir string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer srcFile.Close()

	dstPath := filepath.Join(dstDir, filepath.Base(srcPath))
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer func() {
		dstFile.Close()
		if err != nil {
			os.Remove(dstPath)
		}
	}()

	err = mc.Convert(srcFile, dstFile)
	if err != nil {
		return fmt.Errorf("error converting file: %w", err)
	}
	return nil
}
