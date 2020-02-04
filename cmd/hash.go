package cmd

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	addFlag(hashCmd, "hash-output-file", "e", "public/latest.hash", "The output hash file")
	addFlag(hashCmd, "hash-directory", "d", "public/", "The directory for the hash file")
}

type hashFile struct {
	Version string            `yaml:"version"`
	Files   map[string]string `yaml:"data"`
}

type hashFlags struct {
	OutputFile string `mapstructure:"hash-output-file"`
	Directory  string `mapstructure:"hash-directory"`
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
			data.Version = gitHash("7")
			data.Files = make(map[string]string)

			err := os.Remove(options.OutputFile)
			check(err)

			err = filepath.Walk(options.Directory, func(path string, info os.FileInfo, err error) error {
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
			check(err)

			yaml, err := yaml.Marshal(&data)
			check(err)
			writeToFile(options.OutputFile, yaml)
		},
		TraverseChildren: false,
	}
)

func hash(path string) string {
	f, err := os.Open(path)
	check(err)
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
	check(err)
	return d
}

func gitHash(length string) string {
	if len(length) > 0 {
		return execCmdOutput("git", "rev-parse", "--short="+length, "HEAD")
	}
	return execCmdOutput("git", "rev-parse", "HEAD")
}

func execCmdOutput(name string, arg ...string) string {
	log.Debug(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output:\n", string(out))
	check(err)
	return string(bytes.TrimRight(out, "\n"))
}
