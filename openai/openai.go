package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"wechatbot/config"
)

// proxyURL 代理地址
var proxyURL string

// ProxyClient 客户端
var client http.Client

// ChatComplete 聊天请求
func ChatComplete(msg []map[string]string) string {
	url := "https://api.openai.com/v1/chat/completions"
	data := map[string]interface{}{"model": config.Conf.OpenAPI.Model, "messages": msg, "temperature": config.Conf.OpenAPI.Temperature, "max_tokens": config.Conf.OpenAPI.MaxTokens}
	bytesData, _ := json.Marshal(data)

	return requestAndResponse("application/json", "POST", url, bytes.NewReader(bytesData), func(result map[string]interface{}) string {
		return result["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	})
}

// ImgCreate 图片生成
func ImgCreate(msg string) string {
	url := "https://api.openai.com/v1/images/generations"
	data := map[string]interface{}{"prompt": msg, "n": 1, "size": "512x512"}
	bytesData, _ := json.Marshal(data)

	return requestAndResponse("application/json", "POST", url, bytes.NewReader(bytesData), func(result map[string]interface{}) string {
		return result["data"].([]interface{})[0].(map[string]interface{})["url"].(string)
	})
}

// ImgVariate 图片变体
func ImgVariate(path string) string {
	// 加载图片路径
	url := "https://api.openai.com/v1/images/variations"

	file, _ := os.Open(path)
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "image.jpg")
	io.Copy(part, file)
	writer.WriteField("n", "1")
	writer.WriteField("size", "512x512")
	writer.Close()

	return requestAndResponse(writer.FormDataContentType(), "POST", url, body, func(result map[string]interface{}) string {
		return result["data"].([]interface{})[0].(map[string]interface{})["url"].(string)
	})
}

// requestAndResponse 发送请求和返回处理结果
func requestAndResponse(ContentType string, method string, url string, body io.Reader, f func(result map[string]interface{}) string) string {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return ""
	}
	request.Header.Set("Content-Type", ContentType)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Conf.OpenAPI.Key))
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println("发送请求失败:", err)
		return ""
	}
	defer resp.Body.Close()
	// 读取响应内容
	response, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(response))
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return ""
	}
	// 结果转成map
	var result map[string]interface{}
	json.Unmarshal(response, &result)

	return f(result)
}

func init() {
	proxyURL = fmt.Sprintf("http://%s:%d", config.Conf.ProxyAddr.Ip, config.Conf.ProxyAddr.Port)
	// 解析代理
	proxyURLParsed, _ := url.Parse(proxyURL)
	client = http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURLParsed),
		},
	}
}
