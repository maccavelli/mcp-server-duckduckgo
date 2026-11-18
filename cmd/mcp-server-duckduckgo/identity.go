package main

import (
	"errors"
	"fmt"

	"github.com/maccavelli/go-selfupdate-lib/buildinfo"
	"github.com/spf13/cobra"
)

func newIdentityCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "identity",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Annotations:   map[string]string{skipConfigAnnotation: skipConfigValue},
		RunE: func(*cobra.Command, []string) error {
			if OriginalStdout == nil {
				return errors.New("identity: protocol stdout is not preserved")
			}
			_, err := fmt.Fprintln(OriginalStdout, buildinfo.Identity().String())
			return err
		},
	}
}
