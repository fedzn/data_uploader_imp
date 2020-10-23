package XmlTripConfig

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"path/filepath"
)

type MainTripInfo struct {
	Mileage         float64 `json:"mileage" xml:"mileage,attr"`
	TripDate        string  `json:"tripDate" xml:"trip_date,attr"`
	DataSize        int64   `json:"dataSize" xml:"data_size,attr"`
	ImageCount      int64   `json:"imageCount" xml:"image_count,attr"`
	Start_timestamp int64   `json:"startTimestamp" xml:"start_timestamp,attr"`
	End_timestamp   int64   `json:"endTimestamp" xml:"end_timestamp,attr"`
}

type SubTripInfo struct {
	Name            string `json:"name" xml:"name,attr" `
	Start_timestamp int64  `json:"startTimestamp" xml:"start_timestamp,attr" `
	End_timestamp   int64  `json:"endTimestamp" xml:"end_timestamp,attr" `
	Scene           string `json:"scene" xml:"scene,attr" `
	Description     string `json:"description" xml:"description,attr" `
}

type SubTripData struct {
	SubTrip []SubTripInfo `json:"subTrip" xml:"sub_trip" `
}

type SensorInfo struct {
	Name              string `json:"name" xml:"name,attr" `
	Path              string `json:"path" xml:"path,attr" `
	Protocol          string `json:"protocol" xml:"protocol,attr" `
	Delimiter         string `json:"delimiter" xml:"delimiter,attr" `
	Header            string `json:"header" xml:"header,attr" `
	Compress          bool   `json:"compress" xml:"compress,attr" `
	ConvertWebp       bool   `json:"convertWebp" xml:"convertWebp,attr" `
	Compress_interval uint   `json:"compressInterval" xml:"compress_interval,attr" `
}

type SensorData struct {
	HomeDir string       `json:"homeDir" xml:"home_dir,attr" `
	Sensors []SensorInfo `json:"sensorData" xml:"sensor_data" `
}

type ConfigTripInfo struct {
	MainTrip   MainTripInfo `json:"mainTrip" xml:"main_trip" `
	SubTrips   SubTripData  `json:"subTrips" xml:"sub_trips" `
	SensorData SensorData   `json:"sensorDatas" xml:"sensor_datas" `
	ImageData  SensorData   `json:"imageDatas" xml:"image_datas" `
}

func ParseConfigTripInfo(xml_file string) (*ConfigTripInfo, error) {

	content, err := ioutil.ReadFile(xml_file)
	if err != nil {
		log.Printf("ioutil.ReadFile Error [%s]", err)
		return nil, err
	}

	config_trip_info := ConfigTripInfo{}

	if err := xml.Unmarshal(content, &config_trip_info); err != nil {
		log.Printf("json.Unmarshal [%s] Error : [%v]", content, err)
		return nil, err
	}

	log.Println(config_trip_info)

	return &config_trip_info, nil
}

// 别名对路径字典
type Alias_To_Path = map[string]string

// 路径对别名字典
type Path_To_Alias = map[string]string

type Alias_To_SensorInfo = map[string]SensorInfo

// 获取传感器文件列表
func (c *ConfigTripInfo) GetTripSensorFileList(work_dir string) (Alias_To_Path, Path_To_Alias, Alias_To_SensorInfo) {
	alias2path := make(Alias_To_Path)
	path2alias := make(Path_To_Alias)
	alias2sensorinfo := make(Alias_To_SensorInfo)

	record_dir := filepath.Join(work_dir, c.SensorData.HomeDir)
	image_dir := filepath.Join(work_dir, c.ImageData.HomeDir)

	for _, sensor_info := range c.SensorData.Sensors {
		sensor_sub_path := filepath.Join(record_dir, sensor_info.Path)

		alias2path[sensor_info.Name] = sensor_sub_path
		path2alias[sensor_sub_path] = sensor_info.Name
		alias2sensorinfo[sensor_info.Name] = sensor_info
	}

	for _, sensor_info := range c.ImageData.Sensors {
		sensor_sub_path := filepath.Join(image_dir, sensor_info.Path)

		alias2path[sensor_info.Name] = sensor_sub_path
		path2alias[sensor_sub_path] = sensor_info.Name
		alias2sensorinfo[sensor_info.Name] = sensor_info
	}

	return alias2path, path2alias, alias2sensorinfo
}
