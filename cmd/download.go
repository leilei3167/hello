package cmd

import (
	"downloader/internal/executioner"
	"downloader/pkg/tool"
	"github.com/pkg/errors"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var conc int

func init() {
	rootCmd.AddCommand(downloadCmd) //作为子命令添加到根命令
	downloadCmd.Flags().IntVarP(&conc, "conc", "c", runtime.NumCPU(), "并发线程数,默认cpu数量")
}

var downloadCmd = &cobra.Command{

	Use:     "download",
	Short:   "从url中下载文件",
	Example: os.Args[0] + " download [-c=并发线程数] URL",

	Args: cobra.ExactArgs(1), //只接受1个参数即url
	Run: func(cmd *cobra.Command, args []string) {
		ExitWithError(download(args))
	},
}

//先查询是否存在下载任务
func download(arg []string) error {
	//根据url判断当前是否已有下载文件的文件夹,已有则删除
	if err := tool.CleanTrash(arg[0]); err != nil {
		return errors.WithStack(err)
	}

	return executioner.Do(arg[0], nil, conc)
}
