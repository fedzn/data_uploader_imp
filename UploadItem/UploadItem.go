package UploadItem

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"libs/Parser/ServerConfig"
	"libs/Parser/TaskConfig"
	"libs/Parser/XmlTripConfig"
	Ceph "libs/util/CephClient"
	"libs/util/Util"
	"log"
	"os"
	"sync"
	"time"
)

// type PolicyType uint32

// const (
// 	Sensor_G10      PolicyType = 0
// 	Sensor_CPT      PolicyType = 1
// 	Sensor_DR       PolicyType = 2
// 	Sensor_LM       PolicyType = 3
// 	Sensor_IMU      PolicyType = 4
// 	Sensor_GWM      PolicyType = 5
// 	Sensor_LaneList PolicyType = 6
// 	Sensor_Image    PolicyType = 101
// 	Sensor_ImageLog PolicyType = 102
// )

// func (p PolicyType) String() string {
// 	switch p {
// 	case Sensor_G10:
// 		return "Sensor_G10"
// 	case Sensor_CPT:
// 		return "Sensor_CPT"
// 	case Sensor_DR:
// 		return "Sensor_DR"
// 	case Sensor_LM:
// 		return "Sensor_LM"
// 	case Sensor_IMU:
// 		return "Sensor_IMU"
// 	case Sensor_GWM:
// 		return "Sensor_GWM"
// 	case Sensor_LaneList:
// 		return "Sensor_LaneList"
// 	case Sensor_Image:
// 		return "Sensor_Image"
// 	case Sensor_ImageLog:
// 		return "Sensor_ImageLog"

// 	default:
// 		return "UNKNOWN"
// 	}
// }

// func (p PolicyType) String2() string {
// 	switch p {
// 	case Sensor_G10:
// 		return "G10"
// 	case Sensor_CPT:
// 		return "CPT"
// 	case Sensor_DR:
// 		return "DR"
// 	case Sensor_LM:
// 		return "LM"
// 	case Sensor_IMU:
// 		return "IMU"
// 	case Sensor_GWM:
// 		return "GWM"
// 	case Sensor_LaneList:
// 		return "LANELIST"
// 	case Sensor_Image:
// 		return "IMAGE"
// 	case Sensor_ImageLog:
// 		return "IMAGELOG"

// 	default:
// 		return "UNKNOWN"
// 	}
// }

type UploadItem struct {
	Uuid        string   `json:"uuid" `
	FileKey     string   `json:"fileKey" `
	SubFiles    []string `json:"subFiles" `
	Compress    bool     `json:"compress" `
	ConvertWebp bool     `json:"convertWebp" `
	SensorType  bool     `json:"sensorType" `
	// SensorType  PolicyType `json:"SensorType" `
}

func NewUploadItem(uuid string, key string, subfiles []string, compress bool, convert bool, sensor bool) *UploadItem {

	return &UploadItem{
		Uuid:        uuid,
		FileKey:     key,
		SubFiles:    subfiles,
		Compress:    compress,
		ConvertWebp: convert,
		SensorType:  sensor,
	}
}

// re defined UploadItem
type UploadItems = map[string]*UploadItem
type UploadStats = map[string]bool
type UploadTasks = []string

func FlushUploadItems(items *UploadItems, json_file string) error {
	json_data, err := json.Marshal(items)
	if err != nil {
		log.Printf("json.Marshal Error [%s]", err)
		return err
	}

	if err := ioutil.WriteFile(json_file, json_data, os.ModePerm); err != nil {
		log.Printf("ioutil.WriteFile Error [%s]", err)
		return err
	}

	return nil
}

func ParseUploadItems(json_file string) (*UploadItems, error) {

	content, err := ioutil.ReadFile(json_file)
	if err != nil {
		log.Printf("ioutil.ReadFile Error [%s]", err)
		return nil, err
	}

	items := make(UploadItems)
	if err := json.Unmarshal(content, &items); err != nil {
		log.Printf("json.Unmarshal Error [%s]", err)
		return nil, err
	}

	return &items, nil
}

func FlushUploadStats(items *UploadStats, json_file string) error {
	json_data, err := json.Marshal(items)
	if err != nil {
		log.Printf("json.Marshal Error [%s]", err)
		return err
	}

	if err := ioutil.WriteFile(json_file, json_data, os.ModePerm); err != nil {
		log.Printf("ioutil.WriteFile Error [%s]", err)
		return err
	}

	return nil
}

func ParseUploadStats(json_file string) (*UploadStats, error) {

	content, err := ioutil.ReadFile(json_file)
	if err != nil {
		log.Printf("ioutil.ReadFile Error [%s]", err)
		return nil, err
	}

	items := make(UploadStats)
	if err := json.Unmarshal(content, &items); err != nil {
		log.Printf("json.Unmarshal Error [%s]", err)
		return nil, err
	}

	return &items, nil
}

func GetUploadTasks(json_file string) (*UploadStats, *UploadTasks, error) {
	stats, err := ParseUploadStats(json_file)
	if err != nil {
		log.Printf("ParseUploadStats Error [%s]", err)
		return nil, nil, err
	}

	uuids := make(UploadTasks, 0)
	for uuid, success := range *stats {
		if !success {
			uuids = append(uuids, uuid)
		}
	}

	return stats, &uuids, nil
}

type PostSensorInfo struct {
	Path              string `json:"path" `
	Alias             string `json:"alias" `
	Compress          bool   `json:"compress" `
	Compress_interval uint   `json:"compressInterval" `
	Protocol          string `json:"protocol" `
	Delimiter         string `json:"delimiter" `
	Header            string `json:"header" `
}

func NewPostSensorInfo(path string, alias string, info XmlTripConfig.SensorInfo) *PostSensorInfo {

	// log.Println("Protocol", alias, info.Path, info.Protocol)

	return &PostSensorInfo{
		Path:              path,
		Alias:             alias,
		Compress:          info.Compress,
		Compress_interval: info.Compress_interval,
		Protocol:          info.Protocol,
		Delimiter:         info.Delimiter,
		Header:            info.Header,
	}
}

type PostTripInfo struct {
	Session_uuid string                        `json:"sessionUuid" `
	Task_info    *TaskConfig.TaskInfo          `json:"taskInfo" `
	Trip_info    *XmlTripConfig.ConfigTripInfo `json:"tripInfo" `
	SensorDatas  []*PostSensorInfo             `json:"sensorDatas"`
}

func FlushPostMessageInfo(trip_info *PostTripInfo, out_json string) error {

	json_data, err := json.Marshal(trip_info)
	if err != nil {
		log.Printf("json.Marshal Error [%s]", err)
		return err
	}

	if err := ioutil.WriteFile(out_json, json_data, os.ModePerm); err != nil {
		log.Printf("ioutil.WriteFile Error [%s]", err)
		return err
	}

	return nil
}

func SendPostMessageInfo(url string, json_file string) error {
	content, err := ioutil.ReadFile(json_file)
	if err != nil {
		log.Printf("ioutil.ReadFile Error [%s]", err)
		return err
	}

	// 打印数据
	log.Println(string(content))

	data, err := Util.SendRequest(true, url, string(content))
	if err != nil {
		log.Printf("Util.SendRequest Error [%s]", err)
		return err
	}

	log.Printf("Util.SendRequest Response [%s]", data)

	return nil
}

var (
	chan_task chan *UploadItem = make(chan *UploadItem, 10)
	chan_done chan string      = make(chan string, 10)
	wait_task sync.WaitGroup
	wait_done sync.WaitGroup
)

func Do_Upload_Files(server_info *ServerConfig.Server, threadcount int, work_dir string, task_items_json_file string, task_stats_json_file string, post_infos_json_file string) error {
	log.Println("上传文件开始...")

	items, err := ParseUploadItems(task_items_json_file)
	if err != nil {
		log.Printf("Uploader.Parse_from_Json Error [%s]", err)
		return err
	}

	stats, uuids, err := GetUploadTasks(task_stats_json_file)
	if err != nil {
		log.Printf("Uploader.GetUploadItem_Unsuccess Error [%s]", err)
		return err
	}

	wait_task.Add(1)
	go Do_Produce_Task(items, uuids)

	for i := 0; i < threadcount; i++ {
		wait_task.Add(1)
		tmp_webp_dir := fmt.Sprintf("./tmp/webp_dir_%d", i)
		Util.MakeDir(tmp_webp_dir)
		go Do_Consume_Task(server_info.CEPH, tmp_webp_dir)
	}

	wait_done.Add(1)
	go Do_Done_Task(stats, task_stats_json_file)

	wait_task.Wait()
	close(chan_done)

	wait_done.Wait()

	log.Println("上传文件结束")

	log.Println("发送消息开始...")
	// post_infos_json_file
	if err := SendPostMessageInfo(server_info.DataManager.Url, post_infos_json_file); err != nil {
		log.Printf("Uploader.SendDataManageMessage Error [%s]", err)
		return err
	}

	log.Println("发送消息结束")
	return nil
}

func Do_Done_Task(stats *UploadStats, task_stats_json_file string) error {
	defer wait_done.Done()
	for {
		uuid, ok := <-chan_done
		if !ok {
			break
		}

		// 设置为完成
		(*stats)[uuid] = true
		err := FlushUploadStats(stats, task_stats_json_file)
		if err != nil {
			log.Printf("Uploader.FlushUploadItemsStatus Error [%s]", err)
		}
	}

	return nil
}

func Do_Produce_Task(items *UploadItems, uuids *UploadTasks) error {
	defer wait_task.Done()

	for _, uuid := range *uuids {
		if item, ok := (*items)[uuid]; ok {
			chan_task <- item
		}
	}
	close(chan_task)

	return nil
}

func Do_Consume_Task(ceph_info *ServerConfig.CephInfo, tmp_webp_dir string) error {
	defer wait_task.Done()

	ceph_uploader := Ceph.CreateCephClient(ceph_info)
	for {
		item, ok := <-chan_task
		// log.Printf("recv task [%d]: [%d] [%v]", i, v, ok)
		if !ok {
			break
		}

		start := time.Now()
		var err error = nil
		if item.SensorType {
			err = ceph_uploader.Upload_Single_File(item.SubFiles[0], item.FileKey)
		} else {
			if item.ConvertWebp {
				err = ceph_uploader.Upload_Multi_Image_Files(item.SubFiles, item.FileKey, tmp_webp_dir)
			} else {
				err = ceph_uploader.Upload_Multi_Files(item.SubFiles, item.FileKey)
			}
		}
		//
		cost := time.Since(start)

		if err != nil {
			log.Printf("Upload [failure] [%6.2fs] %s [%s]", cost.Seconds(), item.FileKey, err)
		} else {
			log.Printf("Upload [Success] [%6.2fs] %s", cost.Seconds(), item.FileKey)

			chan_done <- item.Uuid
		}
	}

	return nil
}
