package ServerConfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type CephInfo struct {
	End_point   string `json:"End_point"`
	Access_key  string `json:"Access_key"`
	Secret_key  string `json:"Secret_key"`
	Bucket_name string `json:"Bucket_name"`
}

type InfluxdbInfo struct {
	Host     string `json:"Host"`
	Port     uint   `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	Database string `json:"Database"`
}

type DataManagerInfo struct {
	Url string `json:"Url"`
}

type KafkaInfo struct {
	Host  string `json:"Host"`
	Port  uint   `json:"Port"`
	Topic string `json:"Topic"`
}

type Server struct {
	CEPH        *CephInfo        `json:"CEPH"`
	DataManager *DataManagerInfo `json:"DataManager"`
	InfluxDB    *InfluxdbInfo    `json:"InfluxDB"`
	Kafka       *KafkaInfo       `json:"Kafka"`
}

// 创建 Server
func ParseServerConfig(json_file string) (*Server, error) {

	content, err := ioutil.ReadFile(json_file)
	if err != nil {
		log.Printf("ioutil.ReadFile Error [%s]", err)
		return nil, err
	}

	server := Server{}

	if err := json.Unmarshal(content, &server); err != nil {
		log.Printf("json.Unmarshal [%s] Error [%s]", content, err)
		return nil, err
	}

	return &server, nil
}

func (s *Server) Print() {
	log.Println("Server Config")
	log.Println("  CEPH :")
	log.Printf("    End_point   : [%s]", s.CEPH.End_point)
	log.Printf("    Access_key  : [%s]", s.CEPH.Access_key)
	log.Printf("    Secret_key  : [%s]", s.CEPH.Secret_key)
	log.Printf("    Bucket_name : [%s]", s.CEPH.Bucket_name)
	log.Println("  DataManager :")
	log.Printf("    Url         : [%s]", s.DataManager.Url)
	log.Println("  InfluxDB :")
	log.Printf("    Host        : [%s]", s.InfluxDB.Host)
	log.Printf("    Port        : [%d]", s.InfluxDB.Port)
	log.Printf("    User        : [%s]", s.InfluxDB.User)
	log.Printf("    Password    : [%s]", s.InfluxDB.Password)
	log.Printf("    Database    : [%s]", s.InfluxDB.Database)
	log.Println("  Kafka :")
	log.Printf("    Host        : [%s]", s.Kafka.Host)
	log.Printf("    Port        : [%d]", s.Kafka.Port)
	log.Printf("    Topic       : [%s]", s.Kafka.Topic)
}

func test() {

}
