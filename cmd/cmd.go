package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Main the commands method
func Main(rootCmd *cobra.Command) {
	rootCmd.AddCommand(hashCmd)
	rootCmd.AddCommand(latestCmd)
}

func addFlag(command *cobra.Command, name, shorthand string, value string, usage string) {
	command.Flags().StringP(name, shorthand, value, usage)
	AddViper(command, name)
}

func AddViper(command *cobra.Command, name string) {
	err := viper.BindPFlag(name, command.Flags().Lookup(name))
	if err != nil {
		panic(err)
	}
}