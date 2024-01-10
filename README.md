# h2h: Hexo to Hugo FrontMatter Converter

`h2h` is a CLI tool that facilitates the migration of blogs from Hexo to Hugo and vice versa by converting FrontMatter between the two formats. It takes a directory of Markdown files with Hexo or Hugo FrontMatter, converts them to the other format, and writes the converted files to a specified destination directory.

## Installation

Ensure Go is installed on your system. Use the `go install` command to download and install `h2h`:

```shell
go install github.com/pplmx/h2h
```

## Usage

Specify the source (`--src`) and destination (`--dst`) directories,
the target FrontMatter format (`--format`: "yaml" or "toml"),
and the conversion direction (`--direction`: “hexo2hugo” or “hugo2hexo”).

Example command to convert Hexo FrontMatter to Hugo FrontMatter in YAML format:

```shell
h2h --src /path/to/hexo/posts --dst /path/to/hugo/posts --format yaml --direction hexo2hugo
```

## License

`h2h` is licensed under the MIT
