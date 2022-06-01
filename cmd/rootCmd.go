package cmd

import (
	"github.com/sirupsen/logrus"

	"os"

	"github.com/spf13/cobra"
)

//根命令不执行操作
var rootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "简单多线程下载器",
}

var LogLevel bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&LogLevel, "debug", false, "出现错误是否打印堆栈信息")

	if LogLevel {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug日志开启")
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
