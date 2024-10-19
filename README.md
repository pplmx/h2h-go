# h2h: Hexo to Hugo FrontMatter Converter

`h2h` is a CLI tool that facilitates the migration of blogs from Hexo to Hugo and vice versa by converting FrontMatter between the two formats. It processes a directory of Markdown files with either Hexo or Hugo FrontMatter, converts them to the other format, and writes the converted files to a specified destination directory. By default, it converts from Hexo to Hugo using YAML format.

## Features

- Convert between Hexo and Hugo FrontMatter
- Supports both YAML and TOML formats
- Directional conversion (`hexo2hugo` or `hugo2hexo`)
- Logs all conversion activities to a file for easy debugging and monitoring

## Installation

Ensure Go is installed on your system. Use the following command to download and install `h2h`:

```shell
go install github.com/pplmx/h2h
```

## Usage

### Basic Command

To perform a conversion, specify the source directory (`--src`), destination directory (`--dst`), target FrontMatter format (`--format`: "yaml" or "toml"), and the conversion direction (`--direction`: “hexo2hugo” or “hugo2hexo”).

Example command to convert Hexo FrontMatter to Hugo FrontMatter in YAML format:

```shell
h2h --src /path/to/hexo/posts --dst /path/to/hugo/posts
```

### Options

- `--src`: Source directory containing Markdown files (required)
- `--dst`: Destination directory for converted Markdown files (required)
- `--format`: Target FrontMatter format (`yaml` or `toml`) (default: `yaml`)
- `--direction`: Conversion direction (`hexo2hugo` or `hugo2hexo`) (default: `hexo2hugo`)

### Logging

`h2h` outputs all logs to a file called `h2h.log` in the working directory. This log file contains details of the conversion process, errors, and success messages. This feature is useful for debugging large batch conversions.

### Example Command

Convert from Hugo FrontMatter to Hexo using TOML format:

```shell
h2h --src /path/to/hugo/posts --dst /path/to/hexo/posts --format toml --direction hugo2hexo
```

### Handling Errors

If the conversion fails due to incorrect paths, invalid format, or conversion direction, appropriate error messages will be logged and displayed in the terminal. Check the `h2h.log` file for detailed logs.

## Development

If you would like to contribute or modify the tool, clone the repository and install dependencies using Go:

```shell
git clone https://github.com/pplmx/h2h.git
cd h2h
go build
```

## License

`h2h` is licensed under the MIT License.
