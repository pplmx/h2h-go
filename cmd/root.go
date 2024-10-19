package cmd

import (
	"fmt"
	"github.com/pplmx/h2h/internal"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	defaultFormat    = "yaml"
	defaultDirection = "hexo2hugo"
)

var (
	srcDir       string
	dstDir       string
	targetFormat string
	direction    string
	logFile      *os.File
)

var rootCmd = &cobra.Command{
	Use:   "h2h",
	Short: "Convert between Hexo and Hugo FrontMatter",
	Long: `h2h is a tool to convert between Hexo and Hugo FrontMatter.
It can be used to migrate a Hexo blog to Hugo or a Hugo blog to Hexo.
The tool processes Markdown files with either Hexo or Hugo FrontMatter and converts them to the other format.
Converted files are written to the specified destination directory.

By default, it converts from Hexo to Hugo format using YAML.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("Starting conversion from [%s] to [%s] format, output will be written to [%s]", srcDir, targetFormat, dstDir)
		log.Printf("Conversion direction: %s", direction)

		srcDirAbs, err := filepath.Abs(srcDir)
		if err != nil {
			log.Printf("Error getting absolute path for source directory: %v", err)
			return fmt.Errorf("failed to get absolute path for source directory: %w", err)
		}

		dstDirAbs, err := filepath.Abs(dstDir)
		if err != nil {
			log.Printf("Error getting absolute path for destination directory: %v", err)
			return fmt.Errorf("failed to get absolute path for destination directory: %w", err)
		}

		keyMap, err := getKeyMap(direction)
		if err != nil {
			log.Printf("Invalid direction: %v", err)
			return err
		}

		if err := internal.ConvertPosts(srcDirAbs, dstDirAbs, keyMap, targetFormat); err != nil {
			log.Printf("Conversion failed: %v", err)
			return err
		}

		log.Printf("Conversion completed successfully")
		return nil
	},
}

func Execute() {
	defer cleanup()
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
		os.Exit(1)
	}
}

func init() {
	var err error
	logFile, err = os.OpenFile("h2h.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	log.SetOutput(logFile)

	rootCmd.Flags().StringVar(&srcDir, "src", "", "source directory containing Markdown files to convert (required)")
	rootCmd.Flags().StringVar(&dstDir, "dst", "", "destination directory to write converted Markdown files (required)")
	rootCmd.Flags().StringVar(&targetFormat, "format", defaultFormat, "target FrontMatter format (yaml or toml)")
	rootCmd.Flags().StringVar(&direction, "direction", defaultDirection, "conversion direction (hexo2hugo or hugo2hexo)")

	err = rootCmd.MarkFlagRequired("src")
	cobra.CheckErr(err)

	err = rootCmd.MarkFlagRequired("dst")
	cobra.CheckErr(err)
}

func getKeyMap(direction string) (internal.KeyMap, error) {
	switch direction {
	case "hexo2hugo":
		return internal.HexoToHugoKeyMap, nil
	case "hugo2hexo":
		return internal.HugoToHexoKeyMap, nil
	default:
		return nil, fmt.Errorf("invalid conversion direction: %s", direction)
	}
}

func cleanup() {
	if logFile != nil {
		logFile.Close()
	}
}
