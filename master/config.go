package master

import (
	"encoding/json"
	"io/ioutil"
)

// 配置文件
type Config struct {
	ApiPort int `json:"api_port"`
	ApiReadTimeout int `json:"api_read_timeout"`
	ApiWriteTimeout int `json:"api_write_timeout"`
	EtcdEndpoints []string `json:"etcdEndpoints"`
	EtcdDialTimeout int `json:"etcdDialTimeout"`
}

var (
	G_config *Config
)

func InitConfig(fileName string) (err error) {
	var (
		content []byte
		config Config
	)
	// 读取配置文件
	if content, err = ioutil.ReadFile(fileName); err != nil {
		return err
	}
	// JSON 反序列化
	if err = json.Unmarshal(content, &config); err != nil {
		return err
	}
	// 赋值单例
	G_config = &config
	return nil
}