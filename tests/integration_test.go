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

func createTempFile(t testing.TB, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	require.NoError(t, err, "Failed to create test file %s", name)
}

func verifyFileContent(t *testing.T, dir, name, expectedContent string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	require.NoError(t, err, "Failed to read converted file %s", name)

	assert.Equal(t, 2, strings.Count(string(content), "---"), "Expected 2 '---' separators in %s", name)
	assert.Contains(t, string(content), expectedContent, "Converted file %s does not contain expected content", name)
}

func TestConvertPosts(t *testing.T) {
	testCases := []struct {
		name         string
		files        []struct{ name, content string }
		config       *internal.Config
		expectError  bool
		errorMessage string
	}{
		{
			name: "Basic conversion(Hexo2Hugo)",
			files: []struct{ name, content string }{
				{
					name: "test1.md",
					content: `---
title: Test Post 1
date: 2023-05-01
tags: [test, markdown]
---
# Test Post 1
This is a test post.`,
				},
				{
					name: "test2.md",
					content: `---
title: Test Post 2
date: 2023-05-02
categories: [testing]
---
# Test Post 2
This is a test post.`,
				},
			},
			config:      internal.NewDefaultConfig(),
			expectError: false,
		},
		{
			name: "Invalid front matter",
			files: []struct{ name, content string }{
				{
					name: "invalid.md",
					content: `# Invalid Post
This is an invalid post without front matter.`,
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
			srcDir := t.TempDir()
			dstDir := t.TempDir()

			for _, file := range tc.files {
				createTempFile(t, srcDir, file.name, file.content)
			}

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
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	largeContent := strings.Repeat("This is a large test post.\n", 10000)
	largeFile := `---
title: Large Post
date: 2023-05-01
---
` + largeContent
	createTempFile(t, srcDir, "large.md", largeFile)

	cfg := internal.NewDefaultConfig()
	err := internal.ConvertPosts(srcDir, dstDir, cfg)
	assert.NoError(t, err, "ConvertPosts failed for large file")

	verifyFileContent(t, dstDir, "large.md", "This is a large test post.")
}

func TestConvertNestedDirectories(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	nestedDir := filepath.Join(srcDir, "nested")
	err := os.Mkdir(nestedDir, 0755)
	require.NoError(t, err, "Failed to create nested directory")

	nestedFile := `---
title: Nested Post
date: 2023-05-01
---
# Nested Post
This is a nested post.`
	createTempFile(t, nestedDir, "nested.md", nestedFile)

	cfg := internal.NewDefaultConfig()
	err = internal.ConvertPosts(srcDir, dstDir, cfg)
	assert.NoError(t, err, "ConvertPosts failed for nested directories")

	verifyFileContent(t, filepath.Join(dstDir, "nested"), "nested.md", "This is a nested post.")
}

func TestConvertWithDifferentConcurrency(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	for i := 0; i < 10; i++ {
		content := fmt.Sprintf(`---
title: Test Post %d
date: 2023-05-%02d
---
# Test Post %d
This is test post number %d.`, i, i+1, i, i)
		createTempFile(t, srcDir, fmt.Sprintf("test%d.md", i), content)
	}

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
	srcDir := b.TempDir()
	dstDir := b.TempDir()

	for i := 0; i < 10; i++ {
		content := fmt.Sprintf(`---
title: Bench Post %d
date: 2023-05-%02d
---
# Bench Post %d
%s`, i, i%30+1, i, strings.Repeat("This is a benchmark post.\n", 10))
		createTempFile(b, srcDir, fmt.Sprintf("bench%d.md", i), content)
	}

	cfg := internal.NewDefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := internal.ConvertPosts(srcDir, dstDir, cfg)
		if err != nil {
			b.Fatalf("ConvertPosts failed: %v", err)
		}
	}
}
