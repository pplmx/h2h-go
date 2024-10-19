package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pplmx/h2h/internal"
)

func TestConvertPosts(t *testing.T) {
	// Create temporary directories for source and destination
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create test markdown files
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
		err := os.WriteFile(filepath.Join(srcDir, tf.name), []byte(tf.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.name, err)
		}
	}

	// Run the conversion
	err := internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	if err != nil {
		t.Fatalf("ConvertPosts failed: %v", err)
	}

	// Check if all files were converted
	for _, tf := range testFiles {
		dstPath := filepath.Join(dstDir, tf.name)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("Expected converted file %s does not exist", dstPath)
		}
	}

	// Check the content of converted files
	for _, tf := range testFiles {
		content, err := os.ReadFile(filepath.Join(dstDir, tf.name))
		if err != nil {
			t.Errorf("Failed to read converted file %s: %v", tf.name, err)
			continue
		}

		if count := strings.Count(string(content), "---"); count != 2 {
			t.Errorf("Expected 2 '---' separators in %s, got %d", tf.name, count)
		}

		if !strings.Contains(string(content), "This is a test post.") {
			t.Errorf("Converted file %s does not contain expected content", tf.name)
		}
	}
}

func TestConvertPostsWithErrors(t *testing.T) {
	// Create temporary directories for source and destination
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create a valid test file
	validFile := `---
title: Valid Post
date: 2023-05-01
---
# Valid Post
This is a valid post.`
	err := os.WriteFile(filepath.Join(srcDir, "valid.md"), []byte(validFile), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid test file: %v", err)
	}

	// Create an invalid test file (missing front matter)
	invalidFile := `# Invalid Post
This is an invalid post without front matter.`
	err = os.WriteFile(filepath.Join(srcDir, "invalid.md"), []byte(invalidFile), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid test file: %v", err)
	}

	// Run the conversion
	err = internal.ConvertPosts(srcDir, dstDir, internal.HexoToHugoKeyMap, "yaml")
	if err == nil {
		t.Fatalf("Expected ConvertPosts to return an error, but it didn't")
	}

	// Check if the error message contains the expected information
	if !strings.Contains(err.Error(), "encountered 1 errors during conversion") {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Check if the valid file was converted
	if _, err := os.Stat(filepath.Join(dstDir, "valid.md")); os.IsNotExist(err) {
		t.Errorf("Expected converted file valid.md does not exist")
	}

	// Check if the invalid file was not converted
	if _, err := os.Stat(filepath.Join(dstDir, "invalid.md")); !os.IsNotExist(err) {
		t.Errorf("Invalid file invalid.md should not have been converted")
	}
}
