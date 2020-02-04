package cmd

import (
	"bufio"
	"github.com/jlaffaye/ftp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func init() {
	addPersistentFlag(serverCmd, "host", "s", "", "Ftp server host name (required)")
	addPersistentFlag(serverCmd, "username", "u", "", "The ftp server user (required)")
	addPersistentFlag(serverCmd, "password", "w", "", "The ftp server password (required)")
	addPersistentFlag(serverCmd, "port", "p", "21", "Ftp server port")

	addFlag(deployCmd, "deploy-diff-file", "f", "latest.diff", "The remote hash file")
	addFlag(deployCmd, "deploy-directory", "d", "public/", "The directory for the hash file")
	addFlag(deployCmd, "deploy-path", "a", "/", "Ftp server path")
	serverCmd.AddCommand(deployCmd)

	addFlag(latestCmd, "latest-local-file", "o", "latest.hash", "The output hash file")
	addFlag(latestCmd, "latest-remote-file", "r", "latest.hash", "The remote hash file")
	serverCmd.AddCommand(latestCmd)
}

type serverFlags struct {
	OutputFile string `mapstructure:"latest-local-file"`
	RemoteFile string `mapstructure:"latest-remote-file"`
	DiffFile   string `mapstructure:"deploy-diff-file"`
	Directory  string `mapstructure:"deploy-directory"`
	Path       string `mapstructure:"deploy-path"`
	Host       string `mapstructure:"host"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	Port       string `mapstructure:"port"`
}

var (
	serverCmd = &cobra.Command{
		Use:              "server",
		Short:            "Server operation",
		Long:             `Server operation`,
		TraverseChildren: true,
	}
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the diff to FTP server",
		Long:  `Deploy the web page changes to FTP server`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readServerFlags()
			log.Infof("Start deployment %s", options.DiffFile)
			diff := loadDiffFile(options.DiffFile)

			size := len(diff.Add) + len(diff.Update) + len(diff.Delete) + len(diff.After) + len(diff.Dir)
			log.Infof("[%d/0]", size)
			i := 1

			// open ftp connection
			ftp := connectFtp(options)

			// check directory
			for _, dir := range diff.Dir {
				log.Infof("[%d/%d] MKDIR  %s", size, i, dir)
				i++
				createDirectory(ftp, dir, options.Path)
			}

			// add new files
			for _, file := range diff.Add {
				log.Infof("[%d/%d] ADD    %s", size, i, file)
				i++
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
				err := ftp.Delete(file)
				check(err)
			}
			// after
			for _, file := range diff.After {
				log.Infof("[%d/%d] AFTER  %s", size, i, file)
				i++
				uploadFile(options.Directory, file, ftp)
			}

			closeFtp(ftp)
		},
		TraverseChildren: true,
	}
	latestCmd = &cobra.Command{
		Use:   "latest",
		Short: "Download latest hash file",
		Long:  `Download latest hash file`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readServerFlags()
			log.Infof("Download latest hash file: %s", options.OutputFile)
			ftp := connectFtp(options)

			res, err := ftp.Retr(options.RemoteFile)
			check(err)

			defer res.Close()

			outFile, err := os.Create(options.OutputFile)
			check(err)

			defer outFile.Close()

			_, err = io.Copy(outFile, res)
			check(err)

			closeFtp(ftp)
		},
		TraverseChildren: true,
	}
)

func createDirectory(ftp *ftp.ServerConn, dir, root string) {
	log.Debugf("Create the directory %s", dir)
	dirs := strings.Split(dir, "/")
	for _, d := range dirs {
		err := ftp.ChangeDir(d)
		if err != nil {
			err = ftp.MakeDir(d)
			check(err)
			err = ftp.ChangeDir(d)
		}
	}
	err := ftp.ChangeDir(root)
	check(err)
}

func closeFtp(ftp *ftp.ServerConn) {
	if err := ftp.Quit(); err != nil {
		log.Fatal(err)
	}
}

func connectFtp(options serverFlags) *ftp.ServerConn {
	log.Debugf("Connect to server %s:%s", options.Host, options.Port)
	c, err := ftp.Dial(options.Host+":"+options.Port, ftp.DialWithTimeout(5*time.Second))
	check(err)
	err = c.Login(options.Username, options.Password)
	check(err)
	return c
}

func uploadFile(dir, file string, ftp *ftp.ServerConn) {
	source := dir + file
	log.Debugf("Upload file %s  to ftp file %s", source, file)
	f, err := os.Open(source)
	check(err)
	tmp := bufio.NewReader(f)

	err = ftp.Stor(file, tmp)
	check(err)
}

func readServerFlags() serverFlags {
	result := serverFlags{}
	err := viper.Unmarshal(&result)
	check(err)
	if len(result.Host) == 0 {
		log.Fatal("The FTP server host name is required")
	}
	if len(result.Username) == 0 {
		log.Fatal("The FTP server user name is required")
	}
	if len(result.Password) == 0 {
		log.Fatal("The FTP server user password is required")
	}
	return result
}

func loadDiffFile(filename string) diffFile {
	result := diffFile{}
	log.Debugf("Load diff file %s", filename)
	yamlFile, err := ioutil.ReadFile(filename)
	check(err)
	err = yaml.Unmarshal(yamlFile, &result)
	check(err)
	return result
}
