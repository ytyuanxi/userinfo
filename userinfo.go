package userinfo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"net/http"
	"strings"
	"time"
)

// Config the plugin configuration.

type Config struct {
	UserinfoURL string `json:"userinfoURL,omitempty"`
}

// CreateConfig creates the default plugin configuration.

func CreateConfig() *Config {
	return &Config{}
}

// Example a plugin.

type userInfo struct {
	next        http.Handler
	name        string
	userinfoURL string
}

// New created a new plugin.

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {

	return &userInfo{

		next:        next,
		name:        name,
		userinfoURL: config.UserinfoURL,
	}, nil
}

//curl  -H "Authorization:Bearer vsN-QlQv2zi3SHPk6fOJuSA4_6h14OPiwkk1rcp-xoY.d4m1wrUnPhaXGeS9l-Z_muhdi0nKSpUiDp2dQnDQaYM"  localhost:8082/test2

func (u *userInfo) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	authorization := "no"

	for header, value := range req.Header {
		if header == "Authorization" {
			authorization = value[0]
		}
	}

	if authorization == "no" {
		fmt.Fprintln(rw, "error_description:The request could not be authorized")
		return
	}
	kv := strings.Split(authorization, " ")
	if len(kv) != 2 || kv[0] != "Bearer" {
		fmt.Fprintln(rw, "error_description:The request could not be authorized")
		return
	}

	claims := get(u.userinfoURL, authorization)

	m := make(map[string]string)
	err := json.Unmarshal([]byte(claims), &m)
	if err != nil {
		fmt.Fprintln(rw, "error_description:The request could not be authorized")
		return

	}

	for k, v := range m {
		if k == "sub" {
			req.Header.Set("gridname", v)
		}

	}
	u.next.ServeHTTP(rw, req)

}

// 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func get(url string, token string) string {

	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	request, err := http.NewRequest("GET", url, nil)

	request.Header.Add("Authorization", token)

	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	return result.String()
}
