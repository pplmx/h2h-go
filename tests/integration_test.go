package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pplmx/h2h/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertPosts(t *testing.T) {
	testCases := []struct {
		name         string
		files        []struct{ name, content string }
		config       *internal.Config
		expectError  bool
		errorMessage string
	}{
		{
			name: "Basic conversion (Hexo2Hugo)",
			files: []struct{ name, content string }{
				{
					name:    "test1.md",
					content: createTestContent("Test Post 1", "2023-05-01", []string{"test", "markdown"}, nil, "This is a test post"),
				},
				{
					name:    "test2.md",
					content: createTestContent("Test Post 2", "2023-05-02", nil, []string{"testing"}, "This is a test post"),
				},
			},
			config:      internal.NewDefaultConfig(),
			expectError: false,
		},
		{
			name: "Invalid front matter",
			files: []struct{ name, content string }{
				{
					name:    "invalid.md",
					content: "# Invalid Post\nThis is an invalid post without front matter.",
				},
			},
			config:       internal.NewDefaultConfig(),
			expectError:  true,
			errorMessage: "encountered 1 errors during conversion",
		},
		{
			name: "Empty file",
			files: []struct{ name, content string }{
				{name: "empty.md", content: ""},
			},
			config:       internal.NewDefaultConfig(),
			expectError:  true,
			errorMessage: "encountered 1 errors during conversion",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srcDir, dstDir := createTestEnvironment(t, tc.files)

			err := internal.ConvertPosts(srcDir, dstDir, tc.config)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMessage != "" {
					assert.Contains(t, err.Error(), tc.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				for _, file := range tc.files {
					verifyFileContent(t, dstDir, file.name, "This is a test post")
				}
			}
		})
	}
}

func TestConvertLargeFile(t *testing.T) {
	srcDir, dstDir := createTestEnvironment(t, []struct{ name, content string }{
		{
			name:    "large.md",
			content: createTestContent("Large Post", "2023-05-01", nil, nil, strings.Repeat("This is a large test post.\n", 10000)),
		},
	})

	cfg := internal.NewDefaultConfig()
	err := internal.ConvertPosts(srcDir, dstDir, cfg)
	assert.NoError(t, err, "ConvertPosts failed for large file")

	verifyFileContent(t, dstDir, "large.md", "This is a large test post.")
}

func TestConvertNestedDirectories(t *testing.T) {
	srcDir, dstDir := createTestEnvironment(t, []struct{ name, content string }{
		{
			name:    "nested/nested.md",
			content: createTestContent("Nested Post", "2023-05-01", nil, nil, "# Nested Post\nThis is a nested post."),
		},
	})

	cfg := internal.NewDefaultConfig()
	err := internal.ConvertPosts(srcDir, dstDir, cfg)
	assert.NoError(t, err, "ConvertPosts failed for nested directories")

	verifyFileContent(t, filepath.Join(dstDir, "nested"), "nested.md", "This is a nested post.")
}

func TestConvertWithDifferentConcurrency(t *testing.T) {
	files := make([]struct{ name, content string }, 10)
	for i := 0; i < 10; i++ {
		files[i] = struct{ name, content string }{
			name:    fmt.Sprintf("test%d.md", i),
			content: createTestContent(fmt.Sprintf("Test Post %d", i), fmt.Sprintf("2023-05-%02d", i+1), nil, nil, fmt.Sprintf("# Test Post %d\nThis is test post number %d.", i, i)),
		}
	}

	srcDir, dstDir := createTestEnvironment(t, files)

	concurrencyLevels := []int{1, 2, 4, 8}
	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency%d", concurrency), func(t *testing.T) {
			cfg := internal.NewDefaultConfig()
			cfg.MaxConcurrency = concurrency
			err := internal.ConvertPosts(srcDir, dstDir, cfg)
			assert.NoError(t, err, "ConvertPosts failed with concurrency %d", concurrency)

			for i := 0; i < 10; i++ {
				verifyFileContent(t, dstDir, fmt.Sprintf("test%d.md", i), fmt.Sprintf("This is test post number %d.", i))
			}
		})
	}
}

func BenchmarkConvertPosts(b *testing.B) {
	files := make([]struct{ name, content string }, 10)
	for i := 0; i < 10; i++ {
		files[i] = struct{ name, content string }{
			name:    fmt.Sprintf("bench%d.md", i),
			content: createTestContent(fmt.Sprintf("Bench Post %d", i), fmt.Sprintf("2023-05-%02d", i%30+1), nil, nil, fmt.Sprintf("# Bench Post %d\n%s", i, strings.Repeat("This is a benchmark post.\n", 10))),
		}
	}

	srcDir, dstDir := createTestEnvironment(b, files)

	cfg := internal.NewDefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := internal.ConvertPosts(srcDir, dstDir, cfg)
		if err != nil {
			b.Fatalf("ConvertPosts failed: %v", err)
		}
	}
}

// Helper functions

func createTestEnvironment(t testing.TB, files []struct{ name, content string }) (string, string) {
	t.Helper()
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	for _, file := range files {
		dir := filepath.Dir(filepath.Join(srcDir, file.name))
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err, "Failed to create directory: %s", dir)
		err = os.WriteFile(filepath.Join(srcDir, file.name), []byte(file.content), 0644)
		require.NoError(t, err, "Failed to create test file: %s", file.name)
	}

	return srcDir, dstDir
}

func createTestContent(title, date string, tags []string, categories []string, content string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`---
title: %s
date: %s
`, title, date))
	if len(tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(tags, ", ")))
	}
	if len(categories) > 0 {
		sb.WriteString(fmt.Sprintf("categories: [%s]\n", strings.Join(categories, ", ")))
	}
	sb.WriteString("---\n")

	sb.WriteString(content)
	return sb.String()
}

func verifyFileContent(t *testing.T, dir, name, expectedContent string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	require.NoError(t, err, "Failed to read converted file %s", name)

	assert.Equal(t, 2, strings.Count(string(content), "---"), "Expected 2 '---' separators in %s", name)
	assert.Contains(t, string(content), expectedContent, "Converted file %s does not contain expected content", name)
}
