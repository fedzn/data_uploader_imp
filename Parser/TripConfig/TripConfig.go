package TripConfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
)

type MainTripInfo struct {
	Mileage         float64 `json:"mileage"`
	TripDate        string  `json:"tripDate"`
	DataSize        int64   `json:"dataSize"`
	ImageCount      int64   `json:"imageCount"`
	Start_timestamp int64   `json:"startTimestamp"`
	End_timestamp   int64   `json:"endTimestamp"`
}

type SubTripInfo struct {
	Name            string `json:"name" `
	Start_timestamp int64  `json:"startTimestamp" `
	End_timestamp   int64  `json:"endTimestamp" `
	Scene           string `json:"scene" `
	Description     string `json:"description" `
}

type SensorInfo struct {
	Compress          bool   `json:"compress" `
	ConvertWebp       bool   `json:"convertWebp" `
	Compress_interval uint   `json:"compressInterval" `
	Path              string `json:"path" `
	Protocol          string `json:"protocol" `
	Delimiter         string `json:"delimiter" `
	Header            string `json:"header" `
}

type SensorData struct {
	HomeDir string                `json:"homeDir" `
	Sensors map[string]SensorInfo `json:"sensors" `
}

type ConfigTripInfo struct {
	MainTrip   *MainTripInfo  `json:"mainTrip" `
	SubTrips   []*SubTripInfo `json:"subSrips" `
	SensorData SensorData     `json:"sensorData" `
	ImageData  SensorData     `json:"imageData" `
}

func ParseConfigTripInfo(json_file string) (*ConfigTripInfo, error) {

	content, err := ioutil.ReadFile(json_file)
	if err != nil {
		log.Printf("ioutil.ReadFile Error [%s]", err)
		return nil, err
	}

	config_trip_info := ConfigTripInfo{}

	if err := json.Unmarshal(content, &config_trip_info); err != nil {
		log.Printf("json.Unmarshal [%s] Error : [%v]", content, err)
		return nil, err
	}

	// log.Println(config_trip_info)

	return &config_trip_info, nil
}

// 获取传感器文件列表
func (c *ConfigTripInfo) GetTripSensorFileList(work_dir string) map[string]string {
	sensor_kvs := make(map[string]string)

	record_dir := filepath.Join(work_dir, c.SensorData.HomeDir)
	image_dir := filepath.Join(work_dir, c.ImageData.HomeDir)

	for sensor_name := range c.SensorData.Sensors {
		sensor_info := c.SensorData.Sensors[sensor_name]
		sensor_sub_path := filepath.Join(record_dir, sensor_info.Path)
		sensor_kvs[sensor_name] = sensor_sub_path
	}

	for sensor_name := range c.ImageData.Sensors {
		sensor_info := c.ImageData.Sensors[sensor_name]
		sensor_sub_path := filepath.Join(image_dir, sensor_info.Path)
		sensor_kvs[sensor_name] = sensor_sub_path
	}

	return sensor_kvs
}
