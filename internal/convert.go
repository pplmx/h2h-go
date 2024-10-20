package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

// Config holds the configuration for the conversion process
type Config struct {
	SourceFormat        string
	TargetFormat        string
	FileExtension       string
	MaxConcurrency      int
	ConversionDirection string // New field to specify conversion direction
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		SourceFormat:        "yaml",
		TargetFormat:        "yaml",
		FileExtension:       ".md",
		MaxConcurrency:      4,
		ConversionDirection: "hexo2hugo",
	}
}

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
	sourceFormat string
	targetFormat string
}

// NewFrontMatterConverter creates a new FrontMatterConverter
func NewFrontMatterConverter(cfg *Config) *FrontMatterConverter {
	var keyMap KeyMap
	if cfg.ConversionDirection == "hexo2hugo" {
		keyMap = HexoToHugoKeyMap
	} else {
		keyMap = HugoToHexoKeyMap
	}

	return &FrontMatterConverter{
		keyMap:       keyMap,
		sourceFormat: cfg.SourceFormat,
		targetFormat: cfg.TargetFormat,
	}
}

func (fmc *FrontMatterConverter) Convert(frontMatter string) (string, error) {
	var frontMatterMap map[string]interface{}
	if err := unmarshalFrontMatter(fmc.sourceFormat, []byte(frontMatter), &frontMatterMap); err != nil {
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
	if err := marshalFrontMatter(fmc.targetFormat, &buf, convertedMap); err != nil {
		return "", fmt.Errorf("error marshaling front matter: %w", err)
	}

	return fmt.Sprintf("---\n%s---", buf.String()), nil
}

func unmarshalFrontMatter(format string, data []byte, v interface{}) error {
	switch format {
	case "yaml":
		return yaml.Unmarshal(data, v)
	case "toml":
		return toml.Unmarshal(data, v)
	default:
		return fmt.Errorf("unsupported front matter format: %s", format)
	}
}

func marshalFrontMatter(format string, w io.Writer, v interface{}) error {
	switch format {
	case "yaml":
		encoder := yaml.NewEncoder(w)
		encoder.SetIndent(4)
		return encoder.Encode(v)
	case "toml":
		return toml.NewEncoder(w).Encode(v)
	default:
		return fmt.Errorf("unsupported front matter format: %s", format)
	}
}

// MarkdownConverter handles the conversion of markdown files
type MarkdownConverter struct {
	fmc *FrontMatterConverter
}

// NewMarkdownConverter creates a new MarkdownConverter
func NewMarkdownConverter(cfg *Config) *MarkdownConverter {
	return &MarkdownConverter{
		fmc: NewFrontMatterConverter(cfg),
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
func ConvertPosts(srcDir, dstDir string, cfg *Config) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("error creating destination directory %s: %w", dstDir, err)
	}

	mc := NewMarkdownConverter(cfg)

	var mu sync.Mutex
	var conversionErrors []*ConversionError

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(cfg.MaxConcurrency)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), cfg.FileExtension) {
			g.Go(func() error {
				if err := convertFile(ctx, mc, path, dstDir); err != nil {
					mu.Lock()
					conversionErrors = append(conversionErrors, &ConversionError{SourceFile: path, Err: err})
					mu.Unlock()
				}
				return nil
			})
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking source directory %s: %w", srcDir, err)
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if len(conversionErrors) > 0 {
		for _, err := range conversionErrors {
			fmt.Printf("Error converting file %s: %v\n", err.SourceFile, err.Err)
		}
		return fmt.Errorf("encountered %d errors during conversion", len(conversionErrors))
	}

	return nil
}

func convertFile(ctx context.Context, mc *MarkdownConverter, srcPath, dstDir string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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
