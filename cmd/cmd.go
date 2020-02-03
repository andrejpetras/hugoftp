package cmd

import (
	"bufio"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// Main the commands method
func Main(rootCmd *cobra.Command) {
	rootCmd.AddCommand(hashCmd, latestCmd, diffCmd, deployCmd)
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

func writeToFile(filename string, data []byte) {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriter(file)

	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
	err = w.Flush()
	if err != nil {
		panic(err)
	}
}
