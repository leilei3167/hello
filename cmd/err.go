package cmd

import (
	"fmt"
	"os"
)

//最上层的错误处理,目前是打印堆栈后退出
func ExitWithError(err error) {
	if err != nil {
		if LogLevel {
			fmt.Printf("%+v", err)
			os.Exit(1)
		} else {
			fmt.Printf("%v", err)
			os.Exit(1)
		}

	}
}
