package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func init() {

	flag.StringVar(&ding.Srv, "srv", "", "服务器名字  game_1  center_1等。")
	flag.StringVar(&ding.AlarmPort, "port", "", "监控的端口 。")
	flag.StringVar(&ding.GamePlat, "plat", "", "平台  gcp aws")
	flag.StringVar(&ding.PCip, "ip", "", "本机的ip尾号 。")
	flag.Parse()
}

var ding Ding
var curPath string

// Ding 钉钉消息发送实体
type Ding struct {
	WebHook, Content, Srv, AlarmPort, GamePlat, PCip string
}

type Message struct {
	Content  string `validate:"required"`
	AtPerson []string
	AtAll    bool
}

func main() {
	dir, _ := os.Getwd()
	curPath = dir
	msg := Message{Content: ding.Content}
	msg.AtAll = true
	msg.AtPerson = []string{"18210104695"}

	if "" == ding.Srv || "" == ding.AlarmPort {
		log("srv 或者 port未指定")
		return
	}

	now := time.Now().Format("15:04:05")
	date := time.Now().Format("2006-01-02")

	shellResult := execShell()

	if "" != shellResult {
		msg.Content = fmt.Sprintf("%s %s plat:【 %s 】 ip:【 %s 】 srv:【 %s 】  \n %s", date, now, ding.GamePlat, ding.PCip, ding.Srv, shellResult)
		ding.WebHook = "https://oapi.dingtalk.com/robot/send?access_token=590f3fe91bb6f0559df04721068480efbc0cf3df1ed21d502744f6e0e827e0d5"
		result := ding.Send(msg)
		fmt.Println(result)
	}

}

// SendMessage 发送普通文本消息
func (ding Ding) SendMessage(message Message) Result {
	return ding.Send(message)
}

func (ding Ding) Send(message interface{}) Result {

	var err error

	var paramsMap map[string]interface{}
	paramsMap = convertMessage(message.(Message))

	var buf []byte
	if buf, err = json.Marshal(paramsMap); err != nil {
		return Result{ErrMsg: "marshal message error:" + err.Error()}
	}

	return postMessage(ding.WebHook, string(buf))
}
func convertMessage(m Message) map[string]interface{} {
	var paramsMap = make(map[string]interface{})
	paramsMap["msgtype"] = "text"
	paramsMap["text"] = map[string]string{"content": m.Content}
	paramsMap["at"] = map[string]interface{}{"atMobiles": m.AtPerson, "isAtAll": m.AtAll}
	return paramsMap
}

type Result struct {
	Success bool
	// ErrMsg 错误信息
	ErrMsg string `json:"errmsg"`
	// 错误码
	ErrCode int `json:"errcode"`
}

func postMessage(url string, message string) Result {
	var result Result

	resp, err := http.Post(url, "application/json", strings.NewReader(message))
	if err != nil {
		result.ErrMsg = "post data to api error:" + err.Error()
		return result
	}

	println("message:", message)

	defer resp.Body.Close()
	var content []byte
	if content, err = ioutil.ReadAll(resp.Body); err != nil {
		result.ErrMsg = "read http response body error:" + err.Error()
		return result
	}

	println("response result:", string(content))
	if err = json.Unmarshal(content, &result); err != nil {
		result.ErrMsg = "unmarshal http response body error:" + err.Error()
		return result
	}

	if result.ErrCode == 0 {
		result.Success = true
	}

	return result
}

func execShell() string {
	shFile := fmt.Sprintf(curPath+"/call.sh %s %s", ding.Srv, ding.AlarmPort)
	cmd := exec.Command("sh", "-c", shFile)

	output, err := cmd.Output()
	if err != nil {
		errorResult := fmt.Sprintf("Execute Shell:%s failed with error:%s", shFile, err.Error())
		log(errorResult)
		return errorResult
	}
	result := string(output)
	log("result: " + result)
	return string(output)
}

func log(message interface{}) {
	_, file, line, _ := runtime.Caller(1)

	now := time.Now().Format("15:04:05.000")
	date := time.Now().Format("2006-01-02")

	str := fmt.Sprintf("%s %s %s:%d  %v", date, now, file, line, message)

	fmt.Println(str)
	if curPath != "" {
		fname := fmt.Sprintf("%s/callLog/callAlarm_%s.log", curPath, date)
		logfile, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		defer logfile.Close()
		if err != nil {
			fmt.Printf("%s %s:%d [日志创建错误] %v\r\n", now, file, line, err)
		}
		logfile.WriteString(str + "\r\n")
	}
}
