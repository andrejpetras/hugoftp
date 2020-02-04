package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

func init() {
	addFlag(diffCmd, "diff-file", "i", "latest.diff", "The diff hash file", false)
	addFlag(diffCmd, "diff-new-hash-file", "b", "public/latest.hash", "The local hash file", false)
	addFlag(diffCmd, "diff-old-hash-file", "e", "latest.hash", "The remote hash file", false)
}

type diffFile struct {
	OldVersion string   `yaml:"oldVersion"`
	NewVersion string   `yaml:"newVersion"`
	Add        []string `yaml:"add"`
	Update     []string `yaml:"update"`
	Delete     []string `yaml:"delete"`
	After      []string `yaml:"after"`
}

type diffFlags struct {
	OutputFile  string `mapstructure:"diff-file"`
	NewHashFile string `mapstructure:"diff-new-hash-file"`
	OldHashFile string `mapstructure:"diff-old-hash-file"`
}

var (
	diffCmd = &cobra.Command{
		Use:   "diff",
		Short: "Creates diff between two hash files",
		Long:  `Creates diff file between two hash files`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readDiffFlags()
			log.Infof("Diff %s <-> %s output %s", options.OldHashFile, options.NewHashFile, options.OutputFile)

			oldHash := loadHash(options.OldHashFile)
			newHash := loadHash(options.NewHashFile)

			diffFile := diffFile{}
			diffFile.OldVersion = oldHash.Version
			diffFile.NewVersion = newHash.Version

			hashFile := options.NewHashFile
			index := strings.LastIndex(hashFile, "/")
			if index != -1 {
				hashFile = hashFile[index+1:]
			}
			diffFile.After = append(diffFile.After, hashFile)

			for k, v := range newHash.Files {
				h := oldHash.Files[k]
				if len(h) == 0 {
					diffFile.Add = append(diffFile.Add, k)
				} else {
					if h != v {
						diffFile.Update = append(diffFile.Update, k)
					}
				}
			}

			for k := range oldHash.Files {
				vv := newHash.Files[k]
				if len(vv) == 0 {
					diffFile.Delete = append(diffFile.Delete, k)
				}
			}

			log.Infof("Add    %d", len(diffFile.Add))
			log.Infof("Update %d", len(diffFile.Update))
			log.Infof("Delete %d", len(diffFile.Delete))
			log.Infof("After %d", len(diffFile.After))

			yaml, err := yaml.Marshal(&diffFile)
			check(err)
			writeToFile(options.OutputFile, yaml)
		},
		TraverseChildren: false,
	}
)

func readDiffFlags() diffFlags {
	mavenOptions := diffFlags{}
	err := viper.Unmarshal(&mavenOptions)
	check(err)
	return mavenOptions
}

func loadHash(filename string) hashFile {
	remoteHash := hashFile{}

	yamlFile, err := ioutil.ReadFile(filename)
	check(err)

	err = yaml.Unmarshal(yamlFile, &remoteHash)
	check(err)

	return remoteHash
}
