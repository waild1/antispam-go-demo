/*
@Author : yidun_dev
@Date : 2020-10-30
@File : image_asynccheck.go
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
	apiUrl     = "http://as.dun.163.com/v5/image/asyncCheck"
	version    = "v5"
	secretId   = "your_secret_id"   //产品密钥ID，产品标识
	secretKey  = "your_secret_key"  //产品私有密钥，服务端生成签名信息使用，请严格保管，避免泄露
	businessId = "your_business_id" //业务ID，易盾根据产品业务特点分配
)

// 请求易盾接口
func check(params url.Values) *simplejson.Json {
	params["secretId"] = []string{secretId}
	params["businessId"] = []string{businessId}
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
	var images []map[string]string
	image1 := map[string]string{
		"name": "https://nos.netease.com/yidun/2-0-0-a6133509763d4d6eac881a58f1791976.jpg",
		"type": "1",
		"data": "https://nos.netease.com/yidun/2-0-0-a6133509763d4d6eac881a58f1791976.jpg",
		//"callbackUrl": "http://***",  //主动回调地址url, 如果设置了则走主动回调逻辑
	}
	//image2 := map[string]string{
	//	"name": "{\"imageId\": 33451123, \"contentId\": 78978}",
	//	"type": "2",
	//	"data": "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",
	//	//"callbackUrl": "http://***",  //主动回调地址url, 如果设置了则走主动回调逻辑
	//}
	images = append(images, image1)
	jsonString, _ := json.Marshal(images)
	params := url.Values{
		"images": []string{string(jsonString)},
	}

	ret := check(params)

	code, _ := ret.Get("code").Int()
	message, _ := ret.Get("msg").String()
	if code == 200 {
		result, _ := ret.Get("result").Map()
		_, _ = result["dealingCount"].(json.Number).Int64()
		checkImages := result["checkImages"].([]interface{})
		for _, checkImageItem := range checkImages {
			if checkImageMap, ok := checkImageItem.(map[string]interface{}); ok {
				dataId, _ := checkImageMap["dataId"].(string)
				name, _ := checkImageMap["name"].(string)
				taskId, _ := checkImageMap["taskId"].(string)
				fmt.Printf("dataId=%s，name=%s，taskId=%s", dataId, name, taskId)
			}
		}
	} else {
		fmt.Printf("ERROR: code=%d, msg=%s", code, message)
	}
}
