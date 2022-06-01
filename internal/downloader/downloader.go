package downloader

import (
	"crypto/tls"
	"downloader/pkg/tool"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	stdurl "net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	acceptRange   = "Accept-Ranges"  //是否支持多线程下载
	contentLength = "Content-Length" //body大小
)

var (
	client = &http.Client{ //下载用的client
		Transport: tr,
	}

	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //不验证服务端的证书等
	}

	skipTLS = true //TODO:?跳过TLS?
)

//下载器
type HTTPDownloader struct {
	URL            string
	File           string
	Part           int64                //分区数
	Len            int64                //要下载的文件的大小
	IPs            []string             //url对应的ip
	SkipTLS        bool                 //跳过TLS?
	DownloadRanges []tool.DownloadRange //各个分区的状态(下载的进度,路径)
	Resumable      bool                 //是否可以断点续传
}

//根据线程数量和url来构建一个下载器,根据线程数量进行下载任务分片
func NewHTTPDownloader(url string, par int) (*HTTPDownloader, error) {
	resumable := true

	//验证url合法性,获取到URL结构体
	parsed, err := stdurl.Parse(url)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	//根据获取到的url中的host 查询对应的ip
	ips, err := net.LookupIP(parsed.Host)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	//将ips转换成string切片
	ipstr := IPtoIPv4Str(ips)
	logrus.Printf("Downloading IP is: [%s]\n", strings.Join(ipstr, " | ")) //跟ip无关 只是验证

	//构建一个请求 检查能否正常连接,并获取信息
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := TryConnect(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	//检查是否支持多线程下载
	if resp.Header.Get(acceptRange) == "" {
		logrus.Println("此资源不支持多线程下载,将使用单线程模式")
		par = 1
	}
	//获取文件的大小
	size := resp.Header.Get(contentLength)
	if size == "" {
		logrus.Println("此资源不包含Content-Length信息,将使用单线程模式")
		par = 1
		size = "1"        //进度条最少要1
		resumable = false //设置为不支持断点续传
	}
	logrus.Printf("开始下载,线程数:[%d]\n", par)
	//解析文件大小
	len, err := strconv.ParseInt(size, 10, 64) //10进制,int64
	if err != nil {
		return nil, errors.WithStack(err)
	}
	//打印大小
	sizeInMb := float64(len) / (1024 * 1024)
	if size == "1" {
		logrus.Printf("待下载文件总大小: 未知\n")
	} else if sizeInMb < 1024 {
		logrus.Printf("待下载文件总大小: [%.1f MB]\n", sizeInMb)
	} else {
		logrus.Printf("待下载文件总大小: [%.1f GB]\n", sizeInMb/1024)
	}

	downloader := &HTTPDownloader{}
	file := filepath.Base(url) //url最末端作为文件名
	downloader.URL = url
	downloader.File = file
	downloader.Part = int64(par) //分区数量
	downloader.Len = len
	downloader.IPs = ipstr
	downloader.SkipTLS = skipTLS
	downloader.DownloadRanges, err = partCalculate(int64(par),
		len, url)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	downloader.Resumable = resumable
	return downloader, nil
}

func IPtoIPv4Str(ips []net.IP) []string {
	res := []string{}
	for _, ip := range ips {

		if ip.To4() != nil { //只要ipv4格式
			res = append(res, ip.String())
		}

	}
	return res
}

//重试3次
func TryConnect(req *http.Request) (*http.Response, error) {
	for i := 1; i < 4; i++ {
		logrus.Debugf("第%d次连接...\n", i)
		testClient := http.Client{
			Timeout:   time.Second * 4 * time.Duration(i), //4,8,12 指数退避
			Transport: tr,
		}
		resp, err := testClient.Do(req)
		if err != nil {
			logrus.Debugf("第%d次连接失败:%v,执行重试...\n", i, err)
			continue
		} else {
			logrus.Println("连接成功")
			return resp, nil
		}
	}
	return nil, errors.New("无法连接到对应url")

}

func partCalculate(par int64, len int64, url string) ([]tool.DownloadRange, error) { //根据线程数,文件大小分配

	ret := []tool.DownloadRange{}
	//创建下载目录,以及文件
	file := filepath.Base(url)
	folder, err := tool.GetFolderFrom(url) //下载任务目录
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := tool.Mkdir(folder); err != nil {
		return nil, errors.WithStack(err)
	}

	for i := int64(0); i < par; i++ {
		start := (len / par) * i
		end := int64(0)
		if i < par-1 {
			end = (len/par)*(i+1) - 1
		} else {
			end = len
		}
		fname := fmt.Sprintf("%s.part%d", file, i)
		path := filepath.Join(folder, fname) //分区文件的路径
		ret = append(ret, tool.DownloadRange{
			URL:     url,
			Path:    path, //这一部分的绝对路径
			StartAt: start,
			EndAt:   end,
		})
	}
	return ret, nil
}
