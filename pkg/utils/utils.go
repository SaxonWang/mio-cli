package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/hidevopsio/mio-cli/pkg/types"
	"github.com/manifoldco/promptui"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"runtime"
	"strings"
	"time"
	xwebsocket "golang.org/x/net/websocket"
)

func GetInput(label string) (userInput string) {
	validate := func(input string) error {
		if len(input) < 8 {
			return errors.New("Password must have more than 8 characters")
		}
		return nil
	}
	checkName := func(input string) error {
		if label == "Username" {
			if input == "" {
				return errors.New("Please Input username!")
			}
		}
		return nil
	}
	if label == "Password" {
		u := promptui.Prompt{
			Label:    label,
			Mask:     '*',
			Validate: validate,
		}
		userInput, _ = u.Run()
	} else {
		u := promptui.Prompt{
			Label:    label,
			Validate: checkName,
		}
		userInput, _ = u.Run()
	}
	return userInput
}

//定义用户用以HTTP登陆的JSON对象
type LoginAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BaseResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

//通过HTTP请求,启动 pipeline
func Start(start types.PipelineStart, url, token string) error {
	jsonByte, err := json.Marshal(start)

	fmt.Println("PipelineStart",string(jsonByte))

	if err != nil {
		fmt.Println("Login Failed ", err)
		return err
	}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Login Failed ", err)
		return err
	}
	defer resp.Body.Close()
	byteResp, _ := ioutil.ReadAll(resp.Body)
	resData := BaseResponse{}
	if err := json.Unmarshal(byteResp, &resData); err != nil {
		return err
	}

	if resData.Code != 200 {
		fmt.Println("resp", string(byteResp))
		return errors.New("pipeline start filed")
	}
	return nil
}

//通过HTTP登陆，返回Token
func Login(url, username, password string) (token string, err error) {
	myAuth := LoginAuth{Username: username, Password: password}
	jsonByte, err := json.Marshal(myAuth)
	if err != nil {
		fmt.Println("Login Failed ", err)
		return token, err
	}

	myToken := BaseResponse{}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonByte))
	if err == nil {
		defer resp.Body.Close()
		byteResp, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(byteResp, &myToken)
		if err == nil {
			if myToken.Code == 200 {
				token = myToken.Data["token"].(string)
				err = errors.New(myToken.Message)
			} else {
				err = errors.New(myToken.Message)
			}
		}
	} else {
		//隐藏登陆完整URL信息
		errs := strings.Split(err.Error(), ":")
		err = errors.New(errs[len(errs)-1])
		return
	}
	if token == "" {
		return token, errors.New("token get failed")
	}
	err = nil
	return token, err
}

//获取用户HOME目录
func GetHomeDir() (string, error) {
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, nil
	}

	if "windows" == runtime.GOOS {
		fmt.Println("windows")
		return homeWindows()
	}
	return homeUnix()
}

//获取*unix系统家目录，不对外提供服务。给GetHomeDir调用
func homeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}
	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}
	return result, nil
}

//获取Windows系统家目录，不对外提供服务。给GetHomeDir调用
func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}
	return home, nil
}

//检查指定目录或者文件是否存在
func PathExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

//向指定文件路径写入数据
func WriteText(filePath, text string) error {
	//temporaryFile := fmt.Sprintf("./script-%d.sh", time.Now().Unix())
	fileObj, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fileObj.Close()
	if _, err := fileObj.WriteString(text); err == nil {
		return err
	}
	fileObj.Sync()
	return nil
}

//从指定文件路径读数据
func ReadText(filePath string) (string, error) {
	fileObj, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	//if fileObj,err := os.OpenFile(name,os.O_RDONLY,0644); err == nil {
	defer fileObj.Close()
	contents, err := ioutil.ReadAll(fileObj)
	if err != nil {
		return "", err
	}
	result := strings.Replace(string(contents), "\n", "", 1)
	return result, nil
}

func WsLogsOut(host, path, query string) {

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: host, Path: path, RawQuery: query}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

//连接ws服务端并且循环接受数据并打印到控制台
func ClientLoop(url string) error {
	WS, err := xwebsocket.Dial(url, "", "http://localhost/")
	if err != nil {
		fmt.Println("failed to connect websocket", err.Error())
		return err
	}
	defer func() {
		if WS != nil {
			WS.Close()
		}
	}()


	WS.Write([]byte("hello service"))


	var msg = make([]byte, 2048)
	for {
		if n, err := WS.Read(msg); err != nil {
			fmt.Println(err.Error())
			return err
		} else {
			fmt.Print(string(msg[:n]))
		}
	}
}
