package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pplmx/h2h/internal"
)

func createTempFile(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", name, err)
	}
}

func verifyFileContent(t *testing.T, dir, name, expectedContent string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Errorf("Failed to read converted file %s: %v", name, err)
		return
	}

	if count := strings.Count(string(content), "---"); count != 2 {
		t.Errorf("Expected 2 '---' separators in %s, got %d", name, count)
	}

	if !strings.Contains(string(content), expectedContent) {
		t.Errorf("Converted file %s does not contain expected content", name)
	}
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
	if err != nil {
		t.Fatalf("ConvertPosts failed: %v", err)
	}

	for _, tf := range testFiles {
		dstPath := filepath.Join(dstDir, tf.name)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("Expected converted file %s does not exist", dstPath)
		}
		verifyFileContent(t, dstDir, tf.name, "This is a test post.")
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
	if err == nil {
		t.Fatalf("Expected ConvertPosts to return an error, but it didn't")
	}

	if !strings.Contains(err.Error(), "encountered 1 errors during conversion") {
		t.Errorf("Unexpected error message: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dstDir, "valid.md")); os.IsNotExist(err) {
		t.Errorf("Expected converted file valid.md does not exist")
	}

	if _, err := os.Stat(filepath.Join(dstDir, "invalid.md")); !os.IsNotExist(err) {
		t.Errorf("Invalid file invalid.md should not have been converted")
	}
}

func TestConvertEmptyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	createTempFile(t, srcDir, "empty.md", "")

	err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	if err == nil {
		t.Fatalf("Expected ConvertPosts to return an error for empty file, but it didn't")
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

	err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	if err != nil {
		t.Fatalf("ConvertPosts failed for large file: %v", err)
	}

	verifyFileContent(t, dstDir, "large.md", "This is a large test post.")
}

func TestConvertNestedDirectories(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	nestedDir := filepath.Join(srcDir, "nested")
	err := os.Mkdir(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	nestedFile := `---
title: Nested Post
date: 2023-05-01
---
# Nested Post
This is a nested post.`
	createTempFile(t, nestedDir, "nested.md", nestedFile)

	err = internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	if err != nil {
		t.Fatalf("ConvertPosts failed for nested directories: %v", err)
	}

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
		err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, tf.targetFormat)
		if err != nil {
			t.Fatalf("ConvertPosts failed for %s: %v", tf.name, err)
		}

		dstPath := filepath.Join(dstDir, tf.name)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("Expected converted file %s does not exist", dstPath)
		}
		verifyFileContent(t, dstDir, tf.name, "This is a test post with")
	}
}
