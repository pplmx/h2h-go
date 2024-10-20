package cmd

import (
	"fmt"
	"github.com/pplmx/h2h/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var (
	srcDir      string
	dstDir      string
	config      *internal.Config
	logFile     *os.File
	logFileName = "h2h.log"
)

var rootCmd = &cobra.Command{
	Use:   "h2h",
	Short: "Convert between Hexo and Hugo FrontMatter",
	Long: `h2h is a tool to convert between Hexo and Hugo FrontMatter.
It can be used to migrate a Hexo blog to Hugo or a Hugo blog to Hexo.
The tool processes Markdown files with either Hexo or Hugo FrontMatter and converts them to the other format.
Converted files are written to the specified destination directory.

By default, it converts from Hexo to Hugo format using YAML.`,
	RunE: runConversion,
}

func Execute() {
	defer cleanup()
	if err := rootCmd.Execute(); err != nil {
		log.Printf("Command execution failed: %v", err)
		os.Exit(1)
	}
}

func init() {
	config = internal.NewDefaultConfig()
	initLogger()
	initFlags()
}

func initLogger() {
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
}

func initFlags() {
	rootCmd.Flags().StringVar(&srcDir, "src", "", "source directory containing Markdown files to convert (required)")
	rootCmd.Flags().StringVar(&dstDir, "dst", "", "destination directory to write converted Markdown files (required)")
	rootCmd.Flags().StringVar(&config.SourceFormat, "source-format", config.SourceFormat, "source FrontMatter format (yaml or toml)")
	rootCmd.Flags().StringVar(&config.TargetFormat, "target-format", config.TargetFormat, "target FrontMatter format (yaml or toml)")
	rootCmd.Flags().StringVar(&config.FileExtension, "file-extension", config.FileExtension, "file extension for Markdown files")
	rootCmd.Flags().IntVar(&config.MaxConcurrency, "max-concurrency", config.MaxConcurrency, "maximum number of concurrent file conversions")
	rootCmd.Flags().StringVar(&config.ConversionDirection, "direction", config.ConversionDirection, "conversion direction (hexo2hugo or hugo2hexo)")

	cobra.CheckErr(rootCmd.MarkFlagRequired("src"))
	cobra.CheckErr(rootCmd.MarkFlagRequired("dst"))
}

func runConversion(cmd *cobra.Command, args []string) error {
	log.Printf("Starting conversion from [%s] to [%s] format, direction: %s, output will be written to [%s]",
		config.SourceFormat, config.TargetFormat, config.ConversionDirection, dstDir)

	srcDirAbs, dstDirAbs, err := getAbsolutePaths()
	if err != nil {
		return err
	}

	if err := internal.ConvertPosts(srcDirAbs, dstDirAbs, config); err != nil {
		log.Printf("Conversion failed: %v", err)
		return err
	}

	log.Printf("Conversion completed successfully")
	return nil
}

func getAbsolutePaths() (string, string, error) {
	srcDirAbs, err := filepath.Abs(srcDir)
	if err != nil {
		log.Printf("Error getting absolute path for source directory: %v", err)
		return "", "", fmt.Errorf("failed to get absolute path for source directory: %w", err)
	}

	dstDirAbs, err := filepath.Abs(dstDir)
	if err != nil {
		log.Printf("Error getting absolute path for destination directory: %v", err)
		return "", "", fmt.Errorf("failed to get absolute path for destination directory: %w", err)
	}

	return srcDirAbs, dstDirAbs, nil
}

func cleanup() {
	if logFile != nil {
		logFile.Close()
	}
}
