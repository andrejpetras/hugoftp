package cmd

import (
	"crypto/sha1"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	addFlag(hashCmd, "output-file", "o", "public/latest.hash", "The output hash file")
	addFlag(hashCmd, "directory", "d", "public/", "The directory for the hash file")
}

type hashFile struct {
	Version string            `yaml:"version"`
	Files   map[string]string `yaml:"data"`
}

type hashFlags struct {
	OutputFile string `mapstructure:"output-file"`
	Directory  string `mapstructure:"directory"`
}

var (
	hashCmd = &cobra.Command{
		Use:   "hash",
		Short: "Create the hash file",
		Long:  `Create the hash file`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readHashFlags()
			log.Infof("Create hash file: %s for the directory: %s", options.OutputFile, options.Directory)

			data := hashFile{}
			data.Version = "1234"
			data.Files = make(map[string]string)

			err := filepath.Walk(options.Directory, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					file := strings.TrimPrefix(path, options.Directory)
					hash := hash(path)
					log.Debugf("%s -  %s\n", file, hash)
					data.Files[file] = hash
				}
				return nil
			})
			if err != nil {
				log.Panic(err)
			}

			yaml, err := yaml.Marshal(&data)
			if err != nil {
				log.Panic("error: %v", err)
			}
			writeToFile(options.OutputFile, yaml)
		},
		TraverseChildren: true,
	}
)

func hash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	bs := h.Sum(nil)
	return hex.EncodeToString(bs[:])
}

func readHashFlags() hashFlags {
	d := hashFlags{}
	err := viper.Unmarshal(&d)
	if err != nil {
		panic(err)
	}
	return d
}
