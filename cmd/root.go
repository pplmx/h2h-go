package cmd

import (
	"fmt"
	"github.com/pplmx/h2h/internal"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Define global variables to store the values of the command line flags
var srcDir string
var dstDir string
var targetFormat string
var direction string

// Define the root command of the CLI tool
var rootCmd = &cobra.Command{
	Use:   "h2h",
	Short: "A tool to convert Hexo FrontMatter to Hugo FrontMatter, or vice versa",
	Long: `h2h is a tool to convert Hexo FrontMatter to Hugo FrontMatter, or vice versa. 
It can be used to migrate a Hexo blog to Hugo or a Hugo blog to Hexo. 
The tool expects a directory containing Markdown files with either Hexo or Hugo FrontMatter and converts them to the other format. 
The converted files are written to a specified destination directory.`,

	// Define the function that will be executed when the root command is run
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Converting Markdown files in %s to %s format and writing output to %s\n", srcDir, targetFormat, dstDir)

		// Convert the source and destination directories to absolute paths
		srcDirAbs, err := filepath.Abs(srcDir)
		if err != nil {
			log.Fatal(err)
		}

		dstDirAbs, err := filepath.Abs(dstDir)
		if err != nil {
			log.Fatal(err)
		}

		// Select the key map based on the conversion direction
		var keyMap map[string]string
		switch direction {
		case "hexo2hugo":
			keyMap = internal.HEXO_TO_HUGO_KEY_MAP
		case "hugo2hexo":
			keyMap = internal.HUGO_TO_HEXO_KEY_MAP
		default:
			log.Fatalf("Invalid conversion direction: %s", direction)
		}

		// Call the ConvertPosts function from the internal package to perform the conversion
		err = internal.ConvertPosts(srcDirAbs, dstDirAbs, keyMap, targetFormat)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Execute function that runs the root command of the CLI tool
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init function that initializes the command line flags for the root command
func init() {
	rootCmd.Flags().StringVar(&srcDir, "src", "", "source directory containing Markdown files to convert (required)")
	rootCmd.Flags().StringVar(&dstDir, "dst", "", "destination directory to write converted Markdown files (required)")
	rootCmd.Flags().StringVar(&targetFormat, "format", "yaml", "target FrontMatter format (yaml or toml)")
	rootCmd.Flags().StringVar(&direction, "direction", "hexo2hugo", "conversion direction (hexo2hugo or hugo2hexo)")

	err := rootCmd.MarkFlagRequired("src")
	cobra.CheckErr(err)

	err = rootCmd.MarkFlagRequired("dst")
	cobra.CheckErr(err)
}
