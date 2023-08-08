package cmd

import (
	"fmt"
	"github.com/pplmx/h2h/internal"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var srcDir string
var dstDir string
var targetFormat string
var direction string

var rootCmd = &cobra.Command{
	Use:   "h2h",
	Short: "A tool to convert Hexo FrontMatter to Hugo FrontMatter",
	Long: `h2h is a tool to convert Hexo FrontMatter to Hugo FrontMatter. It can be used to migrate
a Hexo blog to Hugo. The tool expects a directory containing Markdown files with Hexo FrontMatter
and converts them to Hugo FrontMatter. The converted files are written to a specified destination directory.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Converting Markdown files in %s to %s format and writing output to %s\n", srcDir, targetFormat, dstDir)

		srcDirAbs, err := filepath.Abs(srcDir)
		if err != nil {
			log.Fatal(err)
		}

		dstDirAbs, err := filepath.Abs(dstDir)
		if err != nil {
			log.Fatal(err)
		}

		var keyMap map[string]string
		if direction == "hexo2hugo" {
			keyMap = internal.HEXO_TO_HUGO_KEY_MAP
		} else if direction == "hugo2hexo" {
			keyMap = internal.HUGO_TO_HEXO_KEY_MAP
		} else {
			log.Fatalf("Invalid conversion direction: %s", direction)
		}

		err = internal.ConvertPosts(srcDirAbs, dstDirAbs, keyMap, targetFormat)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.Flags().StringVar(&srcDir, "src", "", "source directory containing Markdown files to convert (required)")
	err := rootCmd.MarkFlagRequired("src")
	cobra.CheckErr(err)

	rootCmd.Flags().StringVar(&dstDir, "dst", "", "destination directory to write converted Markdown files (required)")
	err = rootCmd.MarkFlagRequired("dst")
	cobra.CheckErr(err)

	rootCmd.Flags().StringVar(&targetFormat, "format", "yaml", "target FrontMatter format (yaml or toml)")

	rootCmd.Flags().StringVar(&direction, "direction", "hexo2hugo", "conversion direction (hexo2hugo or hugo2hexo)")
}
