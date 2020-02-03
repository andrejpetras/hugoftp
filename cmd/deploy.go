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
	"strings"
	"time"
)

func init() {
	addFlag(deployCmd, "diff-file", "f", "latest.diff", "The remote hash file")
	addFlag(deployCmd, "directory", "d", "public/", "The directory for the hash file")
	addFtpServer(deployCmd)
}

type deployFlags struct {
	DiffFile  string `mapstructure:"diff-file"`
	Directory string `mapstructure:"directory"`
	Host      string `mapstructure:"host"`
	Path      string `mapstructure:"path"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Port      string `mapstructure:"port"`
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
			ftp, err := ftp.Dial(options.Host+":"+options.Port, ftp.DialWithTimeout(5*time.Second))
			if err != nil {
				log.Fatal(err)
			}
			err = ftp.Login(options.Username, options.Password)
			if err != nil {
				log.Fatal(err)
			}

			// add new files
			for _, file := range diff.Add {
				log.Infof("[%d/%d] ADD    %s", size, i, file)
				i++
				// create directory
				index := strings.LastIndex(file, "/")
				if index != -1 {
					dir := file[0:index]
					log.Debugf("Create the directory %s for the file %s", dir, file)
					err = ftp.MakeDir(dir)
					if err != nil {
						panic(err)
					}
				}
				// upload file
				uploadFile(options.Directory, file, ftp)
			}
			// update files
			for _, file := range diff.Update {
				log.Infof("[%d/%d] UPDATE %s", size, i, file)
				i++
				uploadFile(options.Directory, file, ftp)
			}
			// delete files
			for _, file := range diff.Delete {
				log.Infof("[%d/%d] DELETE %s", size, i, file)
				i++
				err = ftp.Delete(file)
				if err != nil {
					panic(err)
				}
			}
			// after
			for _, file := range diff.After {
				log.Infof("[%d/%d] AFTER  %s", size, i, file)
				i++
				uploadFile(options.Directory, file, ftp)
			}

			if err := ftp.Quit(); err != nil {
				log.Fatal(err)
			}
		},
		TraverseChildren: true,
	}
)

func uploadFile(dir, file string, ftp *ftp.ServerConn) {
	source := dir + file
	log.Debugf("Upload file %s  to ftp file %s", source, file)
	f, err := os.Open(source)
	if err != nil {
		log.Panic(err)
	}
	tmp := bufio.NewReader(f)

	err = ftp.Stor(file, tmp)
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

func loadDiffFile(filename string) diffFile {
	result := diffFile{}
	log.Debugf("Load diff file %s", filename)
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
