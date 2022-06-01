package tool

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

var (
	SaveFolder  string = ".fdlr/"
	stateFile   string = "state.json"
	linuxHome   string = os.Getenv("HOME")
	windowsHome string
)

func init() {
	var err error
	windowsHome, err = homedir.Dir()
	if err != nil {
		logrus.Fatal(err)
	}
}

//根据url返回一个以url末尾资源命名的绝对路径
func GetFolderFrom(url string) (string, error) {
	var homepath string
	var path string
	var absolutePath string
	if runtime.GOOS == "linux" {
		homepath = linuxHome
	} else {
		homepath = windowsHome
	}
	path = filepath.Join(homepath, SaveFolder) //path就是在home目录中创建一个.fdlr的隐藏文件,用于存放下载的分片数据
	//Abs返回路径的绝对表达式,base将返回一段路径的最后元素
	absolutePath, err := filepath.Abs(filepath.Join(homepath,
		SaveFolder, filepath.Base(url)))
	if err != nil {
		return "", errors.WithStack(err) //只附加堆栈
	}
	logrus.Debugf("%v下 path:%v绝对路径为:%v", runtime.GOOS, path, absolutePath)
	//防止路径遍历攻击
	relative, err := filepath.Rel(path, absolutePath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	logrus.Debug("relative:", relative)
	if strings.Contains(relative, "..") {
		return "", errors.WithStack(errors.New("your download file may have a path traversal attack"))
	}
	return absolutePath, nil
}

func IsFolderExisted(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
