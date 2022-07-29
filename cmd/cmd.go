package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var cmd *cobra.Command

func init() {
	cmd = SetImagesCommands()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Main will take the workload of executing/starting the cli, when the command is passed to it.
func Main() {
	err := execute(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}

// execute will actually execute the cli by taking the arguments passed to cli.
func execute(args []string) error {
	cmd.SetArgs(args)
	_, err := cmd.ExecuteC()
	if err != nil {
		return err
	}
	return nil
}
