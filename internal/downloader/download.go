package downloader

import (
	"downloader/pkg/tool"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
)

// Downloading 开启多线程下载,会根据HTTPDownloader中线程数量来创建下载线程
//参数中只有interruptChan被读取,其余全是写入;doneChan在所有线程结束工作时会写入,fileChan是在某个分区
//被下载完毕后会写入,errChan当任一线程发生错误时写入,stateChan当且仅当收到中断信号时会将各分区下载进度保存并写入
func (d *HTTPDownloader) Downloading(doneChan chan<- bool, fileChan chan<- string,
	errChan chan<- error, interruptChan <-chan bool, stateSaveChan chan<- tool.DownloadRange) {

	var bars []*pb.ProgressBar
	var barpool *pb.Pool
	var err error

	//确保是在终端状态下 才初始化进度条
	if DisappearProgressBar() {
		for i, part := range d.DownloadRanges {
			//将每一个分区的数据和进度条绑定在一起,加入pool中
			newbar := pb.New64(part.EndAt - part.StartAt).SetUnits(pb.U_BYTES_DEC).
				Prefix(fmt.Sprintf("Thread-[%d]", i))
			newbar.ShowSpeed = false
			newbar.ShowBar = false

			bars = append(bars, newbar)
		}
		barpool, err = pb.StartPool(bars...) //开启进度条
		if err != nil {
			errChan <- errors.WithStack(err)
			return
		}
	}

	//开始并发下载
	wg := &sync.WaitGroup{}

	for i, p := range d.DownloadRanges {
		wg.Add(1)
		go func(d *HTTPDownloader, i int64, part tool.DownloadRange) {
			defer wg.Done()
			bar := new(pb.ProgressBar)
			if DisappearProgressBar() {
				bar = bars[i] //取出对应的进度条
			}

			var ranges string
			if part.EndAt != d.Len { //中间部分
				ranges = fmt.Sprintf("bytes=%d-%d", part.StartAt, part.EndAt)

			} else { //最后的一部分
				ranges = fmt.Sprintf("bytes=%d-", part.StartAt)
			}

			req, err := http.NewRequest("GET", d.URL, nil)
			if err != nil {
				errChan <- errors.WithStack(err)
				return
			}
			if d.Part > 1 {
				//设置要下载的部分
				req.Header.Add("Range", ranges)
			}

			resp, err := client.Do(req)
			if err != nil {
				errChan <- errors.WithStack(err)
				return
			}
			defer resp.Body.Close()

			//打开分区文件,写入下载内容
			f, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				errChan <- errors.WithStack(err)
				return
			}
			defer f.Close()

			var writer io.Writer
			if DisappearProgressBar() {
				writer = io.MultiWriter(f, bar) //同时向文件和进度条写入
			} else {
				writer = io.MultiWriter(f)
			}

			current := int64(0) //记录已下载的字节位置
			for {
				select {
				case <-interruptChan: //收到中断,保存当前状态后退出
					stateSaveChan <- tool.DownloadRange{
						URL:     d.URL,
						Path:    part.Path,
						StartAt: current + part.StartAt, //排除已下载的部分
						EndAt:   part.EndAt,
					}
					return //此协程退出
				default:
					//每次下载100字节
					written, err := io.CopyN(writer, resp.Body, 100)
					current += written
					if err != nil {
						if err != io.EOF {
							errChan <- errors.WithStack(err)
							return
						}
						//err=nil时继续循环下载,直到EOF完成下载
						fileChan <- part.Path
						return
					}
				}

			}

		}(d, int64(i), p)

	}
	wg.Wait()

	err = barpool.Stop()
	if err != nil {
		errChan <- errors.WithStack(err)
		return
	}
	doneChan <- true //下载完毕

}

func DisappearProgressBar() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
