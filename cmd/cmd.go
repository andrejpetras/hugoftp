package cmd

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
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

func addPersistentFlag(command *cobra.Command, name, shorthand string, value string, usage string) {
	command.PersistentFlags().StringP(name, shorthand, value, usage)
	err := viper.BindPFlag(name, command.PersistentFlags().Lookup(name))
	check(err)
}

func addFlag(command *cobra.Command, name, shorthand string, value string, usage string) {
	command.Flags().StringP(name, shorthand, value, usage)
	err := viper.BindPFlag(name, command.Flags().Lookup(name))
	check(err)
}

func dirName(path string) (string, int) {
	index := strings.LastIndex(path, "/")
	if index != -1 {
		return path[0:index], index
	}
	return "", -1
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
