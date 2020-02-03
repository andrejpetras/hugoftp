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
	rootCmd.AddCommand(hashCmd, latestCmd, diffCmd, deployCmd)
}

func addFtpServer(command *cobra.Command) {
	addFlag(command, "host", "s", "", "Ftp server host name")
	setFlagRequired(command, "host")
	addFlag(command, "username", "u", "", "The ftp server user")
	setFlagRequired(command, "username")
	addFlag(command, "password", "w", "", "The ftp server password")
	setFlagRequired(command, "password")
	addFlag(command, "port", "p", "21", "Ftp server port")
	addFlag(command, "path", "a", "/", "Ftp server path")
}

func setFlagRequired(command *cobra.Command, name string) {
	err := command.MarkFlagRequired(name)
	if err != nil {
		log.Panic(err)
	}
}

func addFlag(command *cobra.Command, name, shorthand string, value string, usage string) {
	command.Flags().StringP(name, shorthand, value, usage)
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
