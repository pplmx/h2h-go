package cmd

import (
	"fmt"
	"github.com/pplmx/h2h/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultFormat    = "yaml"
	defaultDirection = "hexo2hugo"
	logFileName      = "h2h.log"
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
	rootCmd.Flags().StringVar(&targetFormat, "format", defaultFormat, "target FrontMatter format (yaml or toml)")
	rootCmd.Flags().StringVar(&direction, "direction", defaultDirection, "conversion direction (hexo2hugo or hugo2hexo)")

	cobra.CheckErr(rootCmd.MarkFlagRequired("src"))
	cobra.CheckErr(rootCmd.MarkFlagRequired("dst"))
}

func runConversion(cmd *cobra.Command, args []string) error {
	log.Printf("Starting conversion from [%s] to [%s] format, output will be written to [%s]", srcDir, targetFormat, dstDir)
	log.Printf("Conversion direction: %s", direction)

	srcDirAbs, dstDirAbs, err := getAbsolutePaths()
	if err != nil {
		return err
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
