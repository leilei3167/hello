package executioner

import (
	"downloader/internal/downloader"
	"downloader/internal/merger"
	"downloader/pkg/tool"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

//新创建的下载和恢复下载统一的启动入口,新任务state为nil

func Do(url string, state *tool.State, conc int) error {
	start := time.Now()
	//监听退出
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	isInterrupted := false
	var files []string             //各个分区下载的文件绝对路径
	var parts []tool.DownloadRange //各个部分

	doneChan := make(chan bool, 1)      //用于鉴别所有分区是否结束工作
	fileChan := make(chan string, conc) //收集各个下载完成部分的存储路径,最终append到files中
	errChan := make(chan error, 1)
	stateChan := make(chan tool.DownloadRange, 1) //收集每个分区的下载进度(中断时)
	interruptChan := make(chan bool, conc)        //每个下载线程是否正常中断

	//根据state判断是新任务还是断点续传,创建下载器
	var dlr *downloader.HTTPDownloader
	var err error
	if state == nil {
		//新任务,解析url,构建下载器
		dlr, err = downloader.NewHTTPDownloader(url, conc)
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		//旧任务,根据state恢复一个下载器
		return errors.New("暂不支持断点续传")
	}

	//开始下载
	go dlr.Downloading(doneChan, fileChan, errChan, interruptChan, stateChan)
	//收集结果

	for {
		select {
		case <-signalChan:
			isInterrupted = true
			for conc > 0 {
				interruptChan <- true //通知每个线程中断工作,保存状态后退出(stateChan将收到),所有中断完成将激活DoneChan
				conc--
			}

		case file := <-fileChan:
			files = append(files, file) //收集已经下载完毕的部分的路径

		case err = <-errChan:
			return errors.WithStack(err)

		case partState := <-stateChan: //保存每一份的状态(只有被中断才会激活)
			parts = append(parts, partState)

		case <-doneChan: //收到donechan说明所有有协程正常退出,包括被中断或者完成下载
			if isInterrupted {
				//被中断,将parts的state持久化到文件中
				if dlr.Resumable {
					logrus.Printf("Interrupted, saving state ... \n")
					s := &tool.State{
						URL:            url,
						DownloadRanges: parts,
					}
					if err = s.Save(); err != nil {
						return errors.WithStack(err)
					}
					return nil

				} else {
					logrus.Printf("Interrupted, but downloading url is not resumable, silently die\n")
					return nil
				}

			} else {
				//将下载完毕的文件路径切片,和最终要的文件名传入
				err = merger.MergeFile(files, filepath.Base(url))
				if err != nil {
					return errors.WithStack(err)
				}
				//合并完成之后将临时分区的文件夹删除
				if err := tool.CleanTrash(url); err != nil {
					return errors.WithStack(err)
				}
				logrus.Printf("下载完毕,耗时:%s\n", time.Since(start))
				return nil
			}

		}
	}
}
