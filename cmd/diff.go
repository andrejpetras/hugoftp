package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func init() {
	addFlag(diffCmd, "diff-file", "i", "latest.diff", "The diff hash file")
	addFlag(diffCmd, "new-hash-file", "b", "public/latest.hash", "The local hash file")
	addFlag(diffCmd, "old-hash-file", "e", "latest.hash", "The remote hash file")
}

type DiffFile struct {
	OldVersion string   `yaml:"oldVersion"`
	NewVersion string   `yaml:"newVersion"`
	Add        []string `yaml:"add"`
	Update     []string `yaml:"update"`
	Delete     []string `yaml:"delete"`
	After      []string `yaml:"after"`
}

type diffFlags struct {
	OutputFile  string `mapstructure:"diff-file"`
	NewHashFile string `mapstructure:"new-hash-file"`
	OldHashFile string `mapstructure:"old-hash-file"`
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

			diffFile := DiffFile{}
			diffFile.OldVersion = oldHash.Version
			diffFile.NewVersion = newHash.Version

			diffFile.After = append(diffFile.After, options.NewHashFile)

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

			for k, _ := range oldHash.Files {
				vv := newHash.Files[k]
				if len(vv) == 0 {
					diffFile.Delete = append(diffFile.Delete, k)
				}
			}

			log.Infof("Add    %d", len(diffFile.Add))
			log.Infof("Update %d", len(diffFile.Update))
			log.Infof("Delete %d", len(diffFile.Delete))

			yaml, err := yaml.Marshal(&diffFile)
			if err != nil {
				log.Panic(err)
			}
			writeToFile(options.OutputFile, yaml)
		},
		TraverseChildren: true,
	}
)

func readDiffFlags() diffFlags {
	mavenOptions := diffFlags{}
	err := viper.Unmarshal(&mavenOptions)
	if err != nil {
		panic(err)
	}
	return mavenOptions
}

func loadHash(filename string) HashFile {
	remoteHash := HashFile{}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &remoteHash)
	if err != nil {
		log.Panic(err)
	}
	return remoteHash
}
