package cmd

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// Main the commands method
func Main(rootCmd *cobra.Command) {
	rootCmd.AddCommand(hashCmd, diffCmd, serverCmd)
}

func check(e error) {
	if e != nil {
		log.Panic(e)
	}
}

func addPersistentFlag(command *cobra.Command, name, shorthand string, value string, usage string, required bool) {
	command.PersistentFlags().StringP(name, shorthand, value, usage)
	err := viper.BindPFlag(name, command.PersistentFlags().Lookup(name))
	if err != nil {
		panic(err)
	}
	if required {
		err := command.MarkPersistentFlagRequired(name)
		check(err)
	}
}

func addFlag(command *cobra.Command, name, shorthand string, value string, usage string, required bool) {
	command.Flags().StringP(name, shorthand, value, usage)
	err := viper.BindPFlag(name, command.Flags().Lookup(name))
	check(err)
	if required {
		err := command.MarkFlagRequired(name)
		check(err)
	}
}

func writeToFile(filename string, data []byte) {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	check(err)

	file, err := os.Create(filename)
	check(err)
	w := bufio.NewWriter(file)

	_, err = w.Write(data)
	check(err)
	err = w.Flush()
	check(err)
}
