package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pplmx/h2h/internal"
	"github.com/stretchr/testify/assert"
)

func createTempFile(t testing.TB, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", name, err)
	}
}

func verifyFileContent(t *testing.T, dir, name, expectedContent string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	assert.NoError(t, err, "Failed to read converted file %s", name)

	assert.Equal(t, 2, strings.Count(string(content), "---"), "Expected 2 '---' separators in %s", name)
	assert.Contains(t, string(content), expectedContent, "Converted file %s does not contain expected content", name)
}

func TestConvertPosts(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	testFiles := []struct {
		name    string
		content string
	}{
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
This is another test post.`,
		},
	}

	for _, tf := range testFiles {
		createTempFile(t, srcDir, tf.name, tf.content)
	}

	err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	assert.NoError(t, err, "ConvertPosts failed")

	for _, tf := range testFiles {
		dstPath := filepath.Join(dstDir, tf.name)
		_, err := os.Stat(dstPath)
		assert.NoError(t, err, "Expected converted file %s does not exist", dstPath)
		verifyFileContent(t, dstDir, tf.name, "This is another test post.")
	}
}

func TestConvertPostsWithErrors(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	validFile := `---
title: Valid Post
date: 2023-05-01
---
# Valid Post
This is a valid post.`
	createTempFile(t, srcDir, "valid.md", validFile)

	invalidFile := `# Invalid Post
This is an invalid post without front matter.`
	createTempFile(t, srcDir, "invalid.md", invalidFile)

	err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	assert.Error(t, err, "Expected ConvertPosts to return an error")

	assert.Contains(t, err.Error(), "encountered 1 errors during conversion", "Unexpected error message")

	_, err = os.Stat(filepath.Join(dstDir, "valid.md"))
	assert.NoError(t, err, "Expected converted file valid.md does not exist")

	_, err = os.Stat(filepath.Join(dstDir, "invalid.md"))
	assert.True(t, os.IsNotExist(err), "Invalid file invalid.md should not have been converted")
}

func TestConvertEmptyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	createTempFile(t, srcDir, "empty.md", "")

	err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	assert.Error(t, err, "Expected ConvertPosts to return an error for empty file")
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

	err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	assert.NoError(t, err, "ConvertPosts failed for large file")

	verifyFileContent(t, dstDir, "large.md", "This is a large test post.")
}

func TestConvertNestedDirectories(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	nestedDir := filepath.Join(srcDir, "nested")
	err := os.Mkdir(nestedDir, 0755)
	assert.NoError(t, err, "Failed to create nested directory")

	nestedFile := `---
title: Nested Post
date: 2023-05-01
---
# Nested Post
This is a nested post.`
	createTempFile(t, nestedDir, "nested.md", nestedFile)

	err = internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	assert.NoError(t, err, "ConvertPosts failed for nested directories")

	verifyFileContent(t, filepath.Join(dstDir, "nested"), "nested.md", "This is a nested post.")
}

func TestConvertDifferentFrontMatterFormats(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	testFiles := []struct {
		name         string
		content      string
		targetFormat string
	}{
		{
			name: "test_yaml.md",
			content: `---
title: Test YAML Post
date: 2023-05-01
tags: [test, yaml]
---
# Test YAML Post
This is a test post with YAML front matter.`,
			targetFormat: "yaml",
		},
		{
			name: "test_toml.md",
			content: `+++
title = "Test TOML Post"
date = 2023-05-02
tags = ["test", "toml"]
+++
# Test TOML Post
This is a test post with TOML front matter.`,
			targetFormat: "toml",
		},
	}

	for _, tf := range testFiles {
		createTempFile(t, srcDir, tf.name, tf.content)
	}

	for _, tf := range testFiles {
		t.Run(tf.name, func(t *testing.T) {
			err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, tf.targetFormat)
			assert.NoError(t, err, "ConvertPosts failed for %s", tf.name)

			dstPath := filepath.Join(dstDir, tf.name)
			_, err = os.Stat(dstPath)
			assert.NoError(t, err, "Expected converted file %s does not exist", dstPath)
			verifyFileContent(t, dstDir, tf.name, "This is a test post with")
		})
	}
}

func BenchmarkConvertPosts(b *testing.B) {
	srcDir := b.TempDir()
	dstDir := b.TempDir()

	largeContent := strings.Repeat("This is a large test post.\n", 10000)
	largeFile := `---
title: Large Post
date: 2023-05-01
---
` + largeContent
	createTempFile(b, srcDir, "large.md", largeFile)

	for i := 0; i < b.N; i++ {
		err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
		if err != nil {
			b.Fatalf("ConvertPosts failed: %v", err)
		}
	}
}
