package cmd

import (
	"downloader/internal/executioner"
	"downloader/pkg/tool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func init() {
	rootCmd.AddCommand(resumeCmd)
}

var resumeCmd = &cobra.Command{
	Use:     "resume",
	Short:   "恢复下载",
	Example: os.Args[0] + " resume URL/文件名",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ExitWithError(resumeTask(args[0]))
	},
}

func resumeTask(name string) error {
	task := ""
	//根据url或者文件名,获取到文件名
	if tool.IsValidURL(name) {
		task = filepath.Base(name)
	} else {
		task = name
	}
	//尝试找到状态文件
	state, err := tool.Read(task)
	if err != nil {
		return errors.WithStack(err)
	}

	return executioner.Do(state.URL, state, conc)
}
