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
	SaveFolder = ".fdlr/"
	stateFile  = "state.yaml"
	homePath   string
)

func init() {
	var err error
	homePath, err = homedir.Dir() //根据不同的系统自动寻找HOME目录
	if err != nil {
		logrus.Fatal("获取HOME目录失败,请检查环境变量!", err)
	}
}

// GetFolderFrom 根据url返回一个以url末尾资源命名的绝对路径
func GetFolderFrom(url string) (string, error) {
	var path string
	var absolutePath string

	path = filepath.Join(homePath, SaveFolder) //path就是在home目录中创建一个.fdlr的隐藏文件,用于存放下载的分片数据
	//Abs返回路径的绝对表达式,base将返回一段路径的最后元素(即下载文件的文件夹)
	absolutePath, err := filepath.Abs(filepath.Join(homePath,
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
	return err == nil //true就是存在
}

func CleanTrash(url string) error {

	folder, err := GetFolderFrom(url) //得到绝对路径(要下载项目的)
	if err != nil {
		return errors.WithStack(err)
	}

	if IsFolderExisted(folder) {
		//删除之前的临时文件
		logrus.Println("任务文件已存在,执行清除...")
		if err := os.RemoveAll(folder); err != nil {
			return errors.WithStack(err)
		}

	}
	return nil
}
