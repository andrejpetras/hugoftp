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
	addFlag(diffCmd, "diff-file", "i", "latest.diff", "The diff hash file")
	addFlag(diffCmd, "diff-new-hash-file", "b", "public/latest.hash", "The local hash file")
	addFlag(diffCmd, "diff-old-hash-file", "e", "latest.hash", "The remote hash file")
}

type diffFile struct {
	OldVersion string   `yaml:"oldVersion"`
	NewVersion string   `yaml:"newVersion"`
	Dir        []string `yaml:"dir"`
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

			var newDir []string

			for k, v := range newHash.Files {
				h := oldHash.Files[k]
				if len(h) == 0 {
					newDir = appendDir(newDir, k)
					diffFile.Add = append(diffFile.Add, k)
				} else {
					if h != v {
						diffFile.Update = append(diffFile.Update, k)
					}
				}
			}

			var existingDir []string
			for k := range oldHash.Files {
				existingDir = appendDir(existingDir, k)
				vv := newHash.Files[k]
				if len(vv) == 0 {
					diffFile.Delete = append(diffFile.Delete, k)
				}
			}

			if len(newDir) > 0 {
				tmpDir := []string{newDir[0]}
				for _, d := range newDir {
					b, i := HasPrefix(tmpDir, d)
					if b {
						if i != -1 {
							tmpDir[i] = d
						}
					} else {
						tmpDir = appendDir(tmpDir, d)
					}
				}

				for _, d := range tmpDir {
					if !dirExists(existingDir, d) {
						diffFile.Dir = append(diffFile.Dir, d)
					}
				}
			}

			log.Infof("Dir    %d", len(diffFile.Dir))
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

func HasPrefix(tmpDir []string, dir string) (bool, int) {
	for i, d2 := range tmpDir {
		if strings.HasPrefix(d2, dir) {
			return true, -1
		} else if strings.HasPrefix(dir, d2) {
			return true, i
		}
	}
	return false, -1
}

func dirExists(existingDir []string, dir string) bool {
	for _, w := range existingDir {
		if strings.HasPrefix(w, dir) {
			return true
		}
	}
	return false
}

func appendDir(tmp []string, path string) []string {
	d, i := dirName(path)
	if i != -1 {
		return append(tmp, d)
	}
	return tmp
}

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
