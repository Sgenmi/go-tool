package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	AllowExt =[]string{".png",".jpg",".gif"}
	host = flag.String("host","127.0.0.1","Host")
	port = flag.String("port","7900","Port")
	dir = flag.String("dir","/tmp","File directory")
	token = flag.String("token","123456","connect auth")
)

type Ret struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data RetData `json:"data"`
}
type RetData struct {
	CompletePath string `json:"complete_path"`
	Path string `json:"path"`
}

func getFileName(fileExt string) string  {
	randNum := rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(99999)
	fileName := fmt.Sprintf("%s%d%s", time.Now().Format("20060102150405"), randNum, fileExt)
	return fileName
}
func getFilePath(project string)(string,error)  {
	filePath := *dir + "/"+ project +fmt.Sprintf("%s",time.Now().Format("/2006/01/02/"))
	_,err := os.Stat(filePath)
	if err !=nil || os.IsExist(err)  {
		isMkDir := os.MkdirAll(filePath,0755)
		if isMkDir !=nil {
			return "",isMkDir
		}
	}
	return filePath,nil
}

func checkExt(ext string) bool  {
	var isAllow bool = false
	for _,v := range (AllowExt) {
		if v==ext {
			isAllow = true
			break
		}
	}
	return isAllow
}

func resourcesImg(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(1000000)
	project := r.FormValue("project")
	_token := r.FormValue("token")

	if _token != *token {
		retJson,_ := json.Marshal(Ret{Code:101,Msg:"token error"})
		w.Write(retJson)
		return
	}
	if project == "" {
		retJson,_:=json.Marshal(Ret{Code:102,Msg:"project is empty"})
		w.Write(retJson)
		return
	}
	fileHandle,fileReader,err := r.FormFile("file")
	if err !=nil {
		retJson,_ := json.Marshal(Ret{Code:103,Msg:err.Error()})
		w.Write(retJson)
		return
	}
	//获取文件扩展名
	fileExt := path.Ext(fileReader.Filename)
	if fileExt == "" {
		retJson,_ := json.Marshal(Ret{Code:104,Msg:"fileExt is empty"})
		w.Write(retJson)
		return
	}
	if checkExt(fileExt) != true {
		retJson,_ := json.Marshal(Ret{Code:105,Msg:"fileExt is not allow"})
		w.Write(retJson)
		return
	}

	//创建目录
	fileName := getFileName(fileExt)
	filePath,err := getFilePath(project)
	if err !=nil {
		retJson,_ := json.Marshal(Ret{Code:106,Msg:err.Error()})
		w.Write(retJson)
		return
	}
	fileAllPath := filePath+fileName
	_file,_err :=os.Create(fileAllPath)
	if _err != nil {
		retJson,_ := json.Marshal(Ret{Code:107,Msg:_err.Error()})
		w.Write(retJson)
		return
	}
	_, err = io.Copy(_file,fileHandle)
	if err != nil {
		retJson,_ := json.Marshal(Ret{Code:108,Msg:err.Error()})
		w.Write(retJson)
		return
	}
	defer fileHandle.Close()
	defer _file.Close()
	retJson,_ := json.Marshal(Ret{Code:0,Msg:"",Data:RetData{
		CompletePath:fileAllPath,Path: strings.Replace(fileAllPath,*dir + "/"+ project,"",1) }})
	w.Write(retJson)
	return
}
func init()  {
	flag.Parse()
}
func main()  {
	http.HandleFunc("/upImg", resourcesImg)
	logStr := fmt.Sprintf("post: %s:%s/upImg {project,token[%s],file}",*host,*port,*token)
	fmt.Println(logStr)
	log.Fatalln(http.ListenAndServe(*host+":"+ *port,nil))
}
