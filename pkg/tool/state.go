package tool

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

//代表着下载的状态,URL以及下载的进度
type State struct {
	URL            string
	DownloadRanges []DownloadRange
}

//每一个部分的下载进度
type DownloadRange struct {
	URL     string
	Path    string //路径,包含文件名 /home/lei/.fdlr/xxx/xxx.partx
	StartAt int64  //开始下载的位置
	EndAt   int64  //这部分结束位置
}

//根据绝对路径创建文件夹
func Mkdir(folder string) error {
	//先判断目录是否存在
	if _, err := os.Stat(folder); err != nil {
		if err = os.MkdirAll(folder, 0700); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (state *State) Save() error {
	//先获取当前url的下载任务路径
	folder, err := GetFolderFrom(state.URL)
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("Saving states data in %s\n", folder)

	y, err := json.Marshal(state)
	if err != nil {
		return errors.WithStack(err)
	}

	return os.WriteFile(filepath.Join(folder, stateFile), y, 0644)
}
