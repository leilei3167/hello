package merger

import (
	"downloader/internal/downloader"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/cheggaaa/pb"
	"github.com/pkg/errors"
)

//将下载好的分区文件的路径传入,在当前的工作目录拼接完成
func MergeFile(files []string, out string) error {
	//排序,确保分区顺序
	sort.Strings(files)

	bar := new(pb.ProgressBar)
	//bar.ShowBar=false
	if downloader.DisappearProgressBar() {
		fmt.Printf("开始合并文件 \n")
		bar = pb.StartNew(len(files))
	}

	//创建最终文件名称(默认在当前工作目录中!)
	resFile, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return errors.WithStack(err)
	}

	//遍历每一个下载完的部分,依次写入最终的文件中
	for _, file := range files {
		file, err := os.OpenFile(file, os.O_RDONLY, 0600)
		if err != nil {
			return errors.WithStack(err)
		}
		_, err = io.Copy(resFile, file) //拷贝
		if err != nil {
			return errors.WithStack(err)
		}
		if downloader.DisappearProgressBar() {
			bar.Increment()
		}
	}
	if downloader.DisappearProgressBar() {
		bar.Finish()
	}

	return resFile.Close()
}
