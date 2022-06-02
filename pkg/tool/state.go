package tool

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// State 用于将下载进度存入文件;包含各分区下载的详细情况
type State struct {
	URL            string
	DownloadRanges []DownloadRange
}

// DownloadRange 每一个部分的下载进度
type DownloadRange struct {
	URL     string
	Path    string //路径,包含文件名 /home/lei/.fdlr/xxx/xxx.partx
	StartAt int64  //开始下载的位置
	EndAt   int64  //这部分结束位置
}

// Mkdir 根据绝对路径创建文件夹
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

	y, err := yaml.Marshal(state)
	if err != nil {
		return errors.WithStack(err)
	}

	return os.WriteFile(filepath.Join(folder, stateFile), y, 0644)
}

func Read(task string) (*State, error) {
	file := filepath.Join(HomePath, SaveFolder, task, stateFile)
	logrus.Printf("读取状态:%s", file)

	var err error

	//读取配置文件并解码
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var state *State
	if err = yaml.Unmarshal(bytes, &state); err != nil {
		return nil, errors.WithStack(err)
	}
	return state, nil

}
