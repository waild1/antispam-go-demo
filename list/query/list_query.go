/*
@Author : yidun_dev
@Date : 2022-06-10
@File : list_query.go
@Version : 1.0
@Golang : 1.13.5
@Doc : http://dun.163.com/api.html
*/
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/tjfoc/gmsm/sm3"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	apiUrl    = "http://as.dun.163yun.com/v2/list/pageQuery"
	version   = "v2"
	secretId  = "your_secret_id"  //产品密钥ID，产品标识
	secretKey = "your_secret_key" //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
)

// 请求易盾接口
func check(params url.Values) *simplejson.Json {
	params["secretId"] = []string{secretId}
	params["version"] = []string{version}
	params["timestamp"] = []string{strconv.FormatInt(time.Now().UnixNano()/1000000, 10)}
	params["nonce"] = []string{strconv.FormatInt(rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(10000000000), 10)}
	// params["signatureMethod"] = []string{"SM3"} // 签名方法支持国密SM3，默认MD5
	params["signature"] = []string{genSignature(params)}

	resp, err := http.Post(apiUrl, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))

	if err != nil {
		fmt.Println("调用API接口失败:", err)
		return nil
	}

	defer resp.Body.Close()

	contents, _ := ioutil.ReadAll(resp.Body)
	result, _ := simplejson.NewJson(contents)
	return result
}

// 生成签名信息
func genSignature(params url.Values) string {
	var paramStr string
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		paramStr += key + params[key][0]
	}
	paramStr += secretKey
	if params["signatureMethod"] != nil && params["signatureMethod"][0] == "SM3" {
		sm3Reader := sm3.New()
		sm3Reader.Write([]byte(paramStr))
		return hex.EncodeToString(sm3Reader.Sum(nil))
	} else {
		md5Reader := md5.New()
		md5Reader.Write([]byte(paramStr))
		return hex.EncodeToString(md5Reader.Sum(nil))
	}
}

func main() {
	params := url.Values{
		"pageNum":    []string{"1"},
		"pageSize":   []string{"pageSize"},
		"startTime":  []string{"1598951727666"},
		"endTime":    []string{"1598961727666"},
		"listType":   []string{"2"},
		"entityType": []string{"1"},
	}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		result, _ := ret.Get("result").Map()
		count, _ := result["count"].(json.Number).Int64()
		rows := result["rows"].([]interface{})
		fmt.Printf("count: %d, rows:%s", count, rows)
		//for _, row := range rows {
		//if rowMap, ok := row.(map[string]interface{}); ok {
		//listType, _ := rowMap["listType"].(json.Number).Int64()
		//entityType, _ := rowMap["entityType"].(json.Number).Int64()
		//productId, _ := rowMap["productId"].(json.Number).Int64()
		//targetId, _ := rowMap["targetId"].(json.Number).Int64()
		//entity, _ := rowMap["entity"].(string)
		//releaseTime, _ := rowMap["releaseTime"].(json.Number).Int64()
		//source, _ := rowMap["source"].(json.Number).Int64()
		//spamType, _ := rowMap["spamType"].(json.Number).Int64()
		//}
		//}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
