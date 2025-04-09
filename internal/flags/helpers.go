package flags

import (
	"fmt"
	"os"
	"strings"

	"github.com/fenghaojiang/eth-better-abigen/internal/version"
	"github.com/urfave/cli/v2"
)

// NewApp creates an app with sane defaults.
func NewApp(usage string) *cli.App {
	git, _ := version.VCS()
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = version.WithCommit(git.Commit, git.Date)
	app.Usage = usage
	app.Copyright = "Copyright 2013-2025 The go-ethereum Authors"
	app.Before = func(ctx *cli.Context) error {
		MigrateGlobalFlags(ctx)
		return nil
	}
	return app
}

var migrationApplied = map[*cli.Command]struct{}{}

// This function may become unnecessary when https://github.com/urfave/cli/pull/1245 is merged.
func MigrateGlobalFlags(ctx *cli.Context) {
	var iterate func(cs []*cli.Command, fn func(*cli.Command))
	iterate = func(cs []*cli.Command, fn func(*cli.Command)) {
		for _, cmd := range cs {
			if _, ok := migrationApplied[cmd]; ok {
				continue
			}
			migrationApplied[cmd] = struct{}{}
			fn(cmd)
			iterate(cmd.Subcommands, fn)
		}
	}

	// This iterates over all commands and wraps their action function.
	iterate(ctx.App.Commands, func(cmd *cli.Command) {
		if cmd.Action == nil {
			return
		}

		action := cmd.Action
		cmd.Action = func(ctx *cli.Context) error {
			doMigrateFlags(ctx)
			return action(ctx)
		}
	})
}

func doMigrateFlags(ctx *cli.Context) {
	// Figure out if there are any aliases of commands. If there are, we want
	// to ignore them when iterating over the flags.
	aliases := make(map[string]bool)
	for _, fl := range ctx.Command.Flags {
		for _, alias := range fl.Names()[1:] {
			aliases[alias] = true
		}
	}
	for _, name := range ctx.FlagNames() {
		for _, parent := range ctx.Lineage()[1:] {
			if parent.IsSet(name) {
				// When iterating across the lineage, we will be served both
				// the 'canon' and alias formats of all commands. In most cases,
				// it's fine to set it in the ctx multiple times (one for each
				// name), however, the Slice-flags are not fine.
				// The slice-flags accumulate, so if we set it once as
				// "foo" and once as alias "F", then both will be present in the slice.
				if _, isAlias := aliases[name]; isAlias {
					continue
				}
				// If it is a string-slice, we need to set it as
				// "alfa, beta, gamma" instead of "[alfa beta gamma]", in order
				// for the backing StringSlice to parse it properly.
				if result := parent.StringSlice(name); len(result) > 0 {
					ctx.Set(name, strings.Join(result, ","))
				} else {
					ctx.Set(name, parent.String(name))
				}
				break
			}
		}
	}
}

// CheckExclusive verifies that only a single instance of the provided flags was
// set by the user. Each flag might optionally be followed by a string type to
// specialize it further.
func CheckExclusive(ctx *cli.Context, args ...any) {
	set := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		// Make sure the next argument is a flag and skip if not set
		flag, ok := args[i].(cli.Flag)
		if !ok {
			panic(fmt.Sprintf("invalid argument, not cli.Flag type: %T", args[i]))
		}
		// Check if next arg extends current and expand its name if so
		name := flag.Names()[0]

		if i+1 < len(args) {
			switch option := args[i+1].(type) {
			case string:
				// Extended flag check, make sure value set doesn't conflict with passed in option
				if ctx.String(flag.Names()[0]) == option {
					name += "=" + option
					set = append(set, "--"+name)
				}
				// shift arguments and continue
				i++
				continue

			case cli.Flag:
			default:
				panic(fmt.Sprintf("invalid argument, not cli.Flag or string extension: %T", args[i+1]))
			}
		}
		// Mark the flag if it's set
		if ctx.IsSet(flag.Names()[0]) {
			set = append(set, "--"+name)
		}
	}
	if len(set) > 1 {
		fmt.Fprintf(os.Stderr, "Flags %v can't be used at the same time", strings.Join(set, ", "))
		os.Exit(1)
	}
}
