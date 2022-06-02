package cmd

import (
	"downloader/pkg/tool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	rootCmd.AddCommand(taskCmd)
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "查看当前暂存的任务",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ExitWithError(TaskPrint())
	},
}

func TaskPrint() error {
	//到文件存储目录读取
	downloading, err := os.ReadDir(filepath.Join(tool.HomePath, tool.SaveFolder))
	if err != nil {
		return errors.WithStack(err)
	}

	var folders []string
	for _, d := range downloading {
		if d.IsDir() {
			folders = append(folders, d.Name())
		}
	}
	if len(folders) == 0 {
		logrus.Println("当前没有暂存的任务!")
		return nil
	}
	fodlerstring := strings.Join(folders, "\n")
	logrus.Printf("当前暂存的下载任务:\n")
	logrus.Println(fodlerstring)
	return nil
}
