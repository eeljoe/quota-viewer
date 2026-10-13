// quota-viewer-cli 是无界面探针入口:quota-viewer-cli.exe --json
// 抓一轮额度,把结果以一行合法 JSON 打到标准输出并以 0 退出;不创建任何窗口。
package main

import (
	"os"

	"quota-viewer/internal/probe"
)

func main() {
	os.Exit(probe.Run(os.Args[1:], os.Stdout, os.Stderr, nil))
}
