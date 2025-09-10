package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"reddit2md/internal/api"
	cli2 "reddit2md/internal/cmd"
	"reddit2md/internal/md"
	"time"

	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const version = "0.1.0"

func main() {

	var args cli2.Args

	args.LoadFromEnv()

	cliArgs := []cli.Argument{
		&cli.StringArg{
			Name:        "url",
			Destination: &args.Url,
			UsageText:   "url to first Reddit thread",
		},
	}

	// TODO: add flag for starting post, custom regex for next link
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"o"},
			Value:       "./reddit2md/",
			Usage:       "output directory for markdown and bundled files (path will be created if it doesn't exist)",
			Destination: &args.OutputDir,
		},
		&cli.StringFlag{
			Name:        "bundle",
			Aliases:     []string{"b"},
			Usage:       "optional path to bundled file (path must exist) (supports epub) (e.g. --bundle book.epub)",
			Destination: &args.BundlePath,
		},
		&cli.IntFlag{
			Name:        "delay",
			Aliases:     []string{"d"},
			Value:       2,
			Usage:       "delay between Reddit requests (in seconds)",
			Destination: &args.Delay,
		},
		&cli.IntFlag{
			Name:        "limit",
			Aliases:     []string{"l"},
			Value:       0,
			Usage:       "maximum number of posts to follow (0 for no limit)",
			Destination: &args.Limit,
		},
		&cli.BoolFlag{
			Name:        "clean",
			Aliases:     []string{"c"},
			Value:       false,
			Usage:       "remove markdown files (useful when bundling)",
			Destination: &args.Clean,
		},
		&cli.BoolFlag{
			Name:        "silent",
			Aliases:     []string{"s"},
			Value:       false,
			Usage:       "disable verbose output",
			Destination: &args.Silent,
		},
	}

	cmd := &cli.Command{
		Version: version,
		Description: "This tool converts chains of Reddit posts into a collection of markdown files, " +
			"and can optionally be bundled into an EPUB or related formats. It identifies a link to the next " +
			"post and iteratively follows the chain until there are none left.\n\n" +
			"Environment variables REDDIT_CLIENT_ID and REDDIT_CLIENT_SECRET must be set before use. " +
			"The `pandoc` CLI tool is required for bundling.",
		Usage:     "pack a chain of Reddit posts into md, epub, or pdf",
		Arguments: cliArgs,
		ArgsUsage: "<url>",
		Flags:     flags,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			initLogger(args.Silent)

			defer func(l *zap.Logger) {
				err := l.Sync()
				if err != nil {
					panic(err)
				}
			}(zap.L())

			if err := args.Validate(); err != nil {
				zap.S().Fatalf(fmt.Sprintf("%v (see 'reddit2md --help' for help)", err))
			}

			if args.BundlePath == "" && args.Clean {
				zap.S().Warn("clean flag is set but bundle path is not specified, thus nothing will be output")
			}

			return run(ctx, &args)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		zap.S().Fatalw("error", "error", err)
	}
}

// initLogger initializes the global logger.
func initLogger(silent bool) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = ""
	encoderCfg.LevelKey = ""
	encoderCfg.CallerKey = ""
	encoderCfg.NameKey = "name"
	encoderCfg.MessageKey = "msg"
	encoderCfg.EncodeName = func(s string, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(s + ":")
	}

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel
	})

	stdoutSyncer := zapcore.Lock(os.Stdout)
	stderrSyncer := zapcore.Lock(os.Stderr)

	var cores []zapcore.Core

	cores = append(cores, zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		stderrSyncer,
		highPriority,
	))

	// Add info core if not silent (logs to stdout)
	if !silent {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			stdoutSyncer,
			lowPriority,
		))
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core).Named("reddit2md")
	zap.ReplaceGlobals(logger)
}
func run(ctx context.Context, args *cli2.Args) error {
	log := zap.S()

	client, err := api.NewClient(args.ClientId, args.ClientSecret)

	if err != nil {
		return err
	}

	post, err := client.GetLinkedPost(ctx, args.Url)

	if err != nil {
		return err
	}

	// 1. scrape reddit threads and put into md files in a temp dir

	postNumber := 0
	var mdPaths []string

	for post, err := range post.Iter(ctx, client) {
		if args.Delay > 0 && postNumber > 0 {
			time.Sleep(time.Duration(args.Delay) * time.Second)
		}

		postNumber++

		log.Infof("got post %d: %s", postNumber, post.Title)

		if err != nil {
			log.Errorw("error getting post, stopping early", err)
			break
		}

		path := fmt.Sprintf("%s/post_%02d.md", args.OutputDir, postNumber)

		err := md.WritePostToMarkdown(post, fmt.Sprintf("Post %d", postNumber), path)

		if err != nil {
			log.Errorw("error writing post to markdown, skipping", err)
			continue
		}

		mdPaths = append(mdPaths, path)

		log.Infof("wrote post (%d) to markdown file", postNumber)

		if args.Limit > 0 && postNumber >= args.Limit {
			log.Infof("reached limit of %d posts, stopping", args.Limit)
			break
		}
	}

	// 2. if bundle flag is set, bundle the md files into the specified format

	if args.BundlePath != "" {
		ext := path.Ext(args.BundlePath)

		switch ext {
		case ".epub":
			log.Info("bundling markdown files into epub...")

			epubPath := fmt.Sprintf("%s/%s", args.OutputDir, args.BundlePath)

			err := cli2.BundleMarkdownToEPUB(post.Title, post.Author, mdPaths, epubPath)

			if err != nil {
				return err
			}

			log.Infof("wrote epub to %s", epubPath)

			break
		default:
			log.Warnf("unsupported bundle format: %s, skipping bundling", ext)
		}

	}

	// 3. if clean flag is set, remove the md files

	if args.Clean {
		log.Info("cleaning up markdown files...")

		for _, mdPath := range mdPaths {
			err := os.Remove(mdPath)

			if err != nil {
				log.Errorw("error removing markdown file, skipping", err)
				continue
			}
		}
	}

	log.Info("done!")

	return nil
}
