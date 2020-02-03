package cmd

import (
	"github.com/jlaffaye/ftp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"time"
)

func init() {
	addFlag(latestCmd, "output-file", "o", "latest.hash", "The output hash file")
	addFlag(latestCmd, "remote-file", "r", "latest.hash", "The remote hash file")
}

type latestFlags struct {
	OutputFile string `mapstructure:"output-file"`
	RemoteFile string `mapstructure:"remote-file"`
	Host       string `mapstructure:"host"`
	Path       string `mapstructure:"path"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	Port       string `mapstructure:"port"`
}

var (
	latestCmd = &cobra.Command{
		Use:   "latest",
		Short: "Download latest hash file",
		Long:  `Download latest hash file`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readLatestFlags()
			log.Infof("Download latest hash file: %s", options.OutputFile)
			c, err := ftp.Dial(options.Host + ":" + options.Port, ftp.DialWithTimeout(5*time.Second))
			if err != nil {
				log.Fatal(err)
			}
			err = c.Login(options.Username, options.Password)
			if err != nil {
				log.Fatal(err)
			}

			res, err := c.Retr(options.RemoteFile)
			if err != nil {
				log.Fatal(err)
			}

			defer res.Close()

			outFile, err := os.Create(options.OutputFile)
			if err != nil {
				log.Fatal(err)
			}

			defer outFile.Close()

			_, err = io.Copy(outFile, res)
			if err != nil {
				log.Fatal(err)
			}

			if err := c.Quit(); err != nil {
				log.Fatal(err)
			}
		},
		TraverseChildren: true,
	}
)

func readLatestFlags() latestFlags {
	mavenOptions := latestFlags{}
	err := viper.Unmarshal(&mavenOptions)
	if err != nil {
		panic(err)
	}
	return mavenOptions
}
