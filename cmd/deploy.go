package cmd

import (
	"bufio"
	"github.com/jlaffaye/ftp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

func init() {
	addFlag(deployCmd, "diff-file", "d", "latest.diff", "The remote hash file")
}

type deployFlags struct {
	DiffFile string `mapstructure:"diff-file"`
	Host     string `mapstructure:"host"`
	Path     string `mapstructure:"path"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Port     string `mapstructure:"port"`
}

var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the diff to FTP server",
		Long:  `Deploy the web page changes to FTP server`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readDeployFlags()
			log.Infof("Start deployment %s", options.DiffFile)
			diff := loadDiffFile(options.DiffFile)

			size := len(diff.Add) + len(diff.Update) + len(diff.Delete) + len(diff.After)
			log.Infof("[%d/0]", size)
			i := 1

			// open ftp connection
			ftp, err := ftp.Dial(options.Host + ":" + options.Port, ftp.DialWithTimeout(5*time.Second))
			if err != nil {
				log.Fatal(err)
			}
			err = ftp.Login(options.Username, options.Password)
			if err != nil {
				log.Fatal(err)
			}

			// add new files
			for _, v := range diff.Add {
				// create directory

				// upload file
				log.Infof("[%d/%d] ADD    %s", size, i, v)
				i++
			}
			// update files
			for _, v := range diff.Update {
				log.Infof("[%d/%d] UPDATE %s", size, i, v)
				i++
				uploadFile(v, ftp)
			}
			// delete files
			for _, v := range diff.Delete {
				log.Infof("[%d/%d] DELETE %s", size, i, v)
				i++

				err = ftp.Delete(v)
				if err != nil {
					panic(err)
				}
			}
			// after
			for _, v := range diff.After {
				log.Infof("[%d/%d] AFTER  %s", size, i, v)
				i++
			}

			if err := ftp.Quit(); err != nil {
				log.Fatal(err)
			}
		},
		TraverseChildren: true,
	}
)

func uploadFile(filename string, ftp *ftp.ServerConn)  {
	f, err := os.Open(filename)
	if err != nil {
		log.Panic(err)
	}
	tmp := bufio.NewReader(f)

	err = ftp.Stor(filename, tmp)
	if err != nil {
		panic(err)
	}
}

func readDeployFlags() deployFlags {
	mavenOptions := deployFlags{}
	err := viper.Unmarshal(&mavenOptions)
	if err != nil {
		panic(err)
	}
	return mavenOptions
}

func loadDiffFile(filename string) DiffFile {
	result := DiffFile{}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &result)
	if err != nil {
		log.Panic(err)
	}
	return result
}
