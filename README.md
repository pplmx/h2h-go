# h2h

`h2h` is a command-line interface (CLI) tool that converts Hexo FrontMatter to Hugo FrontMatter, or vice versa. It can
be used to migrate a Hexo blog to Hugo or a Hugo blog to Hexo. The tool expects a directory containing Markdown files
with either Hexo or Hugo FrontMatter and converts them to the other format. The converted files are written to a
specified destination directory.

## Installation

To install `h2h`, you need to have Go installed on your system. Then, you can use the `go get` command to download and
install the tool:

```shell
go get github.com/pplmx/h2h-go
```

## Usage

To use `h2h`, you need to specify the source and destination directories using the `--src` and `--dst` flags,
respectively. You can also specify the target FrontMatter format using the `--format` flag (either "yaml" or "toml") and
the conversion direction using the `--direction` flag (either "hexo2hugo" or "hugo2hexo").

Here is an example command that converts Hexo FrontMatter to Hugo FrontMatter in YAML format:

```shell
h2h --src /path/to/hexo/posts --dst /path/to/hugo/posts --format yaml --direction hexo2hugo
```

## License

This project is licensed under the MIT License.
