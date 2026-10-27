// releaseprep is maintainer tooling, not part of the installed updater.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/htooaunglynn/upmacx/internal/release"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var c release.Config
	flag.StringVar(&c.Root, "root", ".", "clean tagged checkout")
	flag.StringVar(&c.Tag, "tag", "", "exact existing vMAJOR.MINOR.PATCH tag")
	flag.StringVar(&c.Output, "out", "", "new artifact directory (must not exist)")
	flag.StringVar(&c.SourceArchive, "source-archive", "", "exact downloaded GitHub tag source archive")
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "positional arguments are unsupported")
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := release.Prepare(ctx, c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Prepared local artifacts and formula. Native validation and publication remain separate gates.")
}
