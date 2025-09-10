# reddit2md

A CLI tool written in Go to convert chains of Reddit threads into markdown files.
It can also bundle them into an EPUB file for easy reading on e-readers.

Note: This tool is intended for personal use only. Honestly, I made it to make reading
stories on r/HFY easier on my e-reader. Don't distribute the EPUBs without permission from the original authors.

## Installation

TBA. Right now you have to build it from source and run the binary.

## Usage

You must have the `pandoc` CLI tool installed to use the EPUB generation feature. Also create an
API key on [Reddit](https://www.reddit.com/prefs/apps) and set the following environment variables:

```
REDDIT_CLIENT_ID=your_client_id
REDDIT_CLIENT_SECRET=your_client_secret
```

Here's the basic usage:

```bash
reddit2md [global options] <url>

GLOBAL OPTIONS:
   --output string, -o string  output directory for markdown and bundled files (path will be created if it doesn't exist) (default: "./reddit2md/")
   --bundle string, -b string  optional path to bundled file (path must exist) (supports epub) (e.g. --bundle book.epub)
   --delay int, -d int         delay between Reddit requests (in seconds) (default: 2)
   --limit int, -l int         maximum number of posts to follow (0 for no limit) (default: 0)
   --clean, -c                 remove markdown files (useful when bundling into epub or pdf) (default: false)
   --silent, -s                disable verbose output (default: false)
   --help, -h                  show help
   --version, -v               print the version
```


## Example

```bash
reddit2md -b book.epub -c <url to first Reddit thread>
```
_This will create an EPUB file called `book.epub` in the output directory and remove the intermediate markdown files._

## Roadmap

- [x] Basic functionality to convert Reddit threads to markdown
- [x] Support for walking through linked threads
- [x] EPUB generation using `pandoc`
- [ ] Make the CLI more user-friendly (especially the -o flag and -b flag)
- [ ] Add support for threads that are continued in comments
- [ ] Add support for appending top comments/replies to the end of posts.
- [ ] Add support for generating PDFs
- [ ] Test more types of regex for finding link to next thread
    - Right now it only supports any variation of `[Next](<url>)`
- [ ] Add better error handling and logging, particularly for a lack of pandoc
- [ ] Make generated EPUBs nicer
    - This might include using heuristics to determine if post titles have chapter numbers in them. If so,
        we can use that instead of counting posts 1, 2, 3, etc. which is what it does now. 
    - Another might be adding a CLI arg that forces non-heuristic chapter titles, and also allows you to specify a starting chapter number.
    - Another may be adding a custom title and cover image.
    - Add a table of contents. (un)
- I'll think of more stuff eventually...

## Contributing

I'm new to Go, feel free to help me on my journey by opening issues or PRs.