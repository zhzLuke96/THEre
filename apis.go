package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	RangoKit "./Rango/core"
	RangoRouter "./Rango/router"
	"./mmgo"
	"gopkg.in/mgo.v2/bson"
)

const onePageMaxDoc = 5

var docsDB = mmgo.NewCtx("there_app", "docs")

type responesBody struct {
	State   string `json:"state"`
	Message string `json:"msg"`

	Content interface{} `json:"content"`
}

func newResp() *responesBody {
	return &responesBody{
		State:   "init",
		Message: "none",
		Content: nil,
	}
}

func (r *responesBody) JSON() []byte {
	ret, err := json.Marshal(r)
	if err == nil {
		return ret
	}
	return nil
}

func errResp(w RangoKit.ResponseWriteBody, msg, content string) {
	w.Write((&responesBody{
		State:   "Error",
		Message: msg,
		Content: content,
	}).JSON())
}

type DocObj struct {
	Title      string   `json:"title"`
	Index      int      `json:"idx"`
	Password   string   `json:"pwd"`
	Content    string   `json:"content"`
	Keywords   []string `json:"kws"`
	LastModify int64    `json:"lastModify"`
	Summary    string   `json:"summary"`
}

type DocResp struct {
	Title      string   `json:"title"`
	Index      int      `json:"idx"`
	Password   bool     `json:"pwd"`
	Content    string   `json:"content"`
	Keywords   []string `json:"kws"`
	LastModify int64    `json:"lastModify"`
	Summary    string   `json:"summary"`
}

func getDocResp(doc *DocObj) DocResp {
	return DocResp{
		Title:      doc.Title,
		Index:      doc.Index,
		Password:   doc.Password != "",
		Content:    doc.Content,
		LastModify: doc.LastModify,
		Keywords:   doc.Keywords,
		Summary:    doc.Summary,
	}
}

// GET 获取文章信息，全文 标题 修改时间 是否加密 以及关键词
func getDoc(w RangoKit.ResponseWriteBody, r *http.Request) {
	var doc DocObj
	// var doc map[string]string
	resp := newResp()
	reqVars := RangoRouter.Vars(r)

	if idx, err := strconv.Atoi(reqVars["index"]); err != nil {
		// 参数错误
		errResp(w, "URL wrong", err.Error())
		return
	} else {
		if err = docsDB.FindOne(bson.M{"index": idx}, nil, &doc); err != nil {
			// 查询失败
			errResp(w, "DB wrong", err.Error())
			return
		}
		// success
		resp.State = "SUCCESS"
		resp.Message = "query success."
		resp.Content = getDocResp(&doc)
		// resp.Content = doc
	}

	w.Write(resp.JSON())
}

// GET 获取文章列表，用于首页显示
func getDocs(w RangoKit.ResponseWriteBody, r *http.Request) {
	var docs []DocObj
	resp := newResp()
	reqVars := RangoRouter.Vars(r)

	if pageNum, err := strconv.Atoi(reqVars["pageNum"]); err != nil {
		// 参数错误
		errResp(w, "server wrong", err.Error())
		return
	} else {
		if err = docsDB.FindPageReverse(pageNum, onePageMaxDoc, nil, nil, &docs); err != nil {
			// 查询失败
			errResp(w, "DB wrong", err.Error())
			return
		}
		var docsResps []DocResp
		for _, doc := range docs {
			docsResps = append(docsResps, getDocResp(&doc))
		}
		// success
		resp.State = "SUCCESS"
		resp.Message = "query success."
		resp.Content = docsResps
	}

	w.Write(resp.JSON())
}

type NewDocBody struct {
	Title    string `json:"title"`
	Index    string `json:"idx"`
	Password string `json:"pwd"`
	Content  string `json:"content"`
	Summary  string `json:"summary"`
}

func GetKeyWord(text string) (res []string) {
	return res
}

func CreateCaptcha() string {
	return fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))
}

// POST 提交新文章
func addDoc(w RangoKit.ResponseWriteBody, r *http.Request) {
	var nDocBody NewDocBody
	resp := newResp()

	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, &nDocBody); err == nil {
		index := 0
		if nDocBody.Index == "" {
			index, _ = strconv.Atoi(CreateCaptcha())
		} else {
			index, _ = strconv.Atoi(nDocBody.Index)
		}
		nDoc := DocObj{
			Title:    nDocBody.Title,
			Index:    index,
			Password: nDocBody.Password,
			Content:  nDocBody.Content,
			// js 不支持Unix也不支持毫秒级，必须变动一下单位
			LastModify: time.Now().Unix() * 1000,
			Keywords:   GetKeyWord(nDocBody.Content),
			Summary:    nDocBody.Summary,
		}
		if err = docsDB.Insert(nDoc); err != nil {
			// 提交失败
			errResp(w, "server wrong", err.Error())
			return
		}
		// success
		resp.State = "SUCCESS"
		resp.Message = "add success."
		resp.Content = ""
	}
	w.Write(resp.JSON())
}

// POST 删除文章
func delDoc(w RangoKit.ResponseWriteBody, r *http.Request) {
	resp := newResp()
	reqVars := RangoRouter.Vars(r)

	idx, _ := strconv.Atoi(reqVars["index"])
	if err := docsDB.Remove(bson.M{"index": idx}); err != nil {
		// 删除失败
		errResp(w, "server wrong", err.Error())
		return
	}
	// success
	resp.State = "SUCCESS"
	resp.Message = "add success."
	resp.Content = ""
	w.Write(resp.JSON())
}

// POST 删除文章
func delDocTitle(w RangoKit.ResponseWriteBody, r *http.Request) {
	resp := newResp()
	reqVars := RangoRouter.Vars(r)

	title := reqVars["title"]
	if err := docsDB.Remove(bson.M{"title": title}); err != nil {
		// 删除失败
		errResp(w, "server wrong", err.Error())
		return
	}
	// success
	resp.State = "SUCCESS"
	resp.Message = "remove success."
	resp.Content = ""
	w.Write(resp.JSON())
}

func getAllKeywords() ([]string, error) {
	var docs []bson.M
	if err := docsDB.FindAll(nil, nil, &docs); err != nil {
		return nil, err
	}
	var ret []string
	for _, doc := range docs {
		if kws, ok := doc["keywords"].([]string); ok {
			for _, k := range kws {
				ret = append(ret, k)
			}
		}
	}
	return ret, nil
}
