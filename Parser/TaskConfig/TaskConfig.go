package TaskConfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Vector3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type AdmsInfo struct {
	Id                string   `json:"id"`
	Coordinate_system string   `json:"coordinateSystem"`
	Enable_bias       bool     `json:"enableBias"`
	Bias              Vector3D `json:"bias"`
	LLH               Vector3D `json:"llh"`
}

type TaskInfo struct {
	Trip_name    string   `json:"tripName" `
	Trip_date    string   `json:"tripDate" `
	Vehicle      string   `json:"vehicle"`
	City         string   `json:"city"`
	Algo_version string   `json:"algoVersion"`
	Device_code  int      `json:"deviceCode"`
	Platform     string   `json:"platform"`
	AdmsInfo     AdmsInfo `json:"admsInfo"`
}

func ParseTaskConfig(json_file string) (*TaskInfo, error) {

	content, err := ioutil.ReadFile(json_file)
	if err != nil {
		log.Printf("ioutil.ReadFile Error [%s]", err)
		return nil, err
	}

	task_info := TaskInfo{}

	if err := json.Unmarshal(content, &task_info); err != nil {
		log.Printf("json.Unmarshal [%s] Error : [%v]", content, err)
		return nil, err
	}

	// log.Println(task_info)

	return &task_info, nil
}
