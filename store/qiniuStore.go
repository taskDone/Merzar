package store

import (
	"fmt"
	"os"
	"qiniupkg.com/api.v7/conf"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
	"strings"
)

type QiniuStore struct{}

var (
	bucket     = "meizi"
	access_key = "1Bhy_R2cAvMp-UHaPhHUz8_u32g6rWwPIWDjX0wu"
	secret_key = "9RsS8LRwVE9hQEpK9fju_GXjUoHlklKLPRyKobKb"

	mysql_host = ""
	mysql_user = ""
	mysql_pwd  = ""
	mysql_db   = ""
)

type PutRet struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

func init() {
	conf.ACCESS_KEY = access_key
	conf.SECRET_KEY = secret_key
}

func (p *QiniuStore) Upload(url string) {
	fmt.Println(url)
	arr := strings.Split(url, "/")
	key := arr[len(arr)-1]

	c := kodo.New(0, nil)
	policy := &kodo.PutPolicy{
		Scope:   bucket + ":" + key,
		Expires: 3600,
	}

	token := c.MakeUptoken(policy)
	zone := 0
	uploader := kodocli.NewUploader(zone, nil)

	var ret PutRet

	res := uploader.PutFile(nil, &ret, token, key, url, nil)
	if res != nil {
		fmt.Println(url, "upload failed ", res)
	}

	os.Remove(url)
}
