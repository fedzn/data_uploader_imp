package main

import (
	"errors"
	"flag"
	"fmt"
	"libs/Parser/ServerConfig"
	"libs/Parser/TaskConfig"
	"libs/Parser/XmlTripConfig"
	"libs/UploadItem"
	"libs/util/Util"
	"log"
	"path/filepath"
	"strings"
)

func Check_Sensor_Files() error {
	// 主要查看设置的文件是否存在
	log.Println("文件检查开始...")

	// 别名 是否存在 大小 文件数量 0文件数量 路径
	log.Printf("%-10s %-10s %-10s %-10s %-10s %-10s", "FileAlias", "Exist", "Size", "Count", "NullFile", "FilePath")

	un_exist_files := make([]string, 0)
	for sensor_alias, sensor_file := range sensors_alias2path {
		is_exist, file_size, file_count, null_file_count := Util.GetSensorFileInfo(sensor_file)
		if !is_exist {
			un_exist_files = append(un_exist_files, sensor_file)
		}
		log.Printf("%-10s %-10t %-10d %-10d %-10d %s", sensor_alias, is_exist, file_size, file_count, null_file_count, sensor_file)
	}

	var err error = nil

	// 判断文件是否有丢失
	if len(un_exist_files) > 0 {
		log.Printf("下列文件丢失...数量:[%d]", len(un_exist_files))
		for idx, sub_file := range un_exist_files {
			log.Printf("%-6d %s", idx, sub_file)
		}
		err = errors.New("file not found")
	} else {
		// 计算cpt里程
		{
			log.Println("开始计算CPT里程")
			cpt_file := sensors_alias2path["CPT"]
			distance, err := Util.Calc_cpt_distance(cpt_file, trip_info.MainTrip.Start_timestamp, trip_info.MainTrip.End_timestamp)
			if err != nil {
				log.Printf("Util.Calc_cpt_distance Error [%s]", err)
				return err
			}
			trip_mileage = distance
		}

		// 计算文件大小
		{
			log.Println("开始计算文件大小")
			record_dir := filepath.Join(work_dir, trip_info.SensorData.HomeDir)
			trip_datasize, _ = Util.Calc_Folder_Capacity(record_dir)

			// 解析行程配置文件，从服务器获取
			// {
			// 	// task_info_file := filepath.Join(record_dir, "trip_config.bin")
			// 	task_info_file := filepath.Join(work_dir, "trip_config.json")
			// 	task_info, err = TaskConfig.ParseTaskConfig(task_info_file)
			// 	if err != nil {
			// 		log.Printf("TaskInfo.ParseTaskInfo Error [%s]", err)
			// 		return err
			// 	}
			// }

			// 行程日期
			log.Println("开始计算行程日期")
			// trip_tripdate = Util.Timestamp_to_date(task_info.Trip_date)
			// trip_tripdate = task_info.Trip_date

			log.Println("开始计算文件统计")
			image_dir := filepath.Join(work_dir, trip_info.ImageData.HomeDir)
			for _, sensor_info := range trip_info.ImageData.Sensors {
				sensor_name := sensor_info.Name
				sensor_sub_dir := filepath.Join(image_dir, sensor_info.Path)

				// 计算图片数量
				log.Println("开始计算图片数量")
				if sensor_name == "IMAGE" {
					img_files := Util.Get_all_files(sensor_sub_dir, true)
					trip_imagecount = int64(len(img_files))
				}

				log.Println("开始计算子文件大小", sensor_sub_dir)
				data_size2, _ := Util.Calc_Folder_Capacity(sensor_sub_dir)
				trip_datasize += data_size2
			}
		}
	}

	// 打印统计信息
	log.Println("开始打印统计信息")
	log.Printf("trip_tripdate   : [%s]", trip_tripdate)
	log.Printf("trip_mileage(m) : [%.3f]", trip_mileage)
	log.Printf("trip_datasize   : [%d]", trip_datasize)
	log.Printf("trip_imagecount : [%d]", trip_imagecount)

	log.Println("文件检查结束")

	return err
}

// 创建上传文件列表
func Create_Upload_Files_Json() error {
	log.Println("创建任务文件开始...")

	// 输出数据列表定义
	items := make(UploadItem.UploadItems)
	stats := make(UploadItem.UploadStats)
	posts := []*UploadItem.PostSensorInfo{}

	// trip_date := Util.Timestamp_to_date(task_info.Trip_date)
	trip_date := task_info.Trip_date
	trip_uuid := Util.Create_uuid()
	platform := task_info.Platform

	record_dir := filepath.Join(work_dir, trip_info.SensorData.HomeDir)
	image_dir := filepath.Join(work_dir, trip_info.ImageData.HomeDir)

	// 生成传感器数据的上传列表文件和状态列表文件
	record_files := []string{}
	// img_files := []string{}
	log.Println(record_dir)
	Util.Get_All_Sub_Files(record_dir, &record_files)
	log.Println(record_files)
	for _, record_file := range record_files {
		// for _, record_file := range Util.Get_all_files(record_dir, false) {

		// 获取传感器文件别名
		sensor_alias := ""
		if alias, ok := sensors_path2alias[record_file]; ok {
			sensor_alias = alias
		}

		// 获取文件的相对路径
		relative_path := strings.TrimPrefix(record_file, work_dir)
		// file_key := fmt.Sprintf("/%s/%s/%s%s", trip_date, trip_uuid, platform, relative_path)
		// 使用/开头的路径，java会有问题,使用./开通也不行
		file_key := fmt.Sprintf("%s/%s/%s%s", trip_date, trip_uuid, platform, relative_path)
		sub_files := []string{record_file}
		uuid := Util.Create_uuid()

		// log.Printf("%-10s -> %-40s -> %s", sensor_alias, relative_path, record_file)

		// 设置需要发送的传感器别名数据
		if sensor_alias != "" {
			sensor_info := sensors_alias2sensorinfo[sensor_alias]
			posts = append(posts, UploadItem.NewPostSensorInfo(file_key, sensor_alias, sensor_info))
		}

		// 任务列表
		items[uuid] = UploadItem.NewUploadItem(uuid, file_key, sub_files, false, false, true)

		// 状态列表
		stats[uuid] = false
	}

	// 生成图像数据的上传列表文件和状态列表文件
	for _, sensor_info := range trip_info.ImageData.Sensors {
		sensor_name := sensor_info.Name

		// 获取文件列表
		sensor_sub_dir := filepath.Join(image_dir, sensor_info.Path)
		sensor_sub_files := Util.Get_all_files(sensor_sub_dir, true)

		// 设置需要发送的图像别名数据
		file_key := fmt.Sprintf("%s/%s/%s/%s", trip_date, trip_uuid, sensor_name, sensor_info.Path)
		posts = append(posts, UploadItem.NewPostSensorInfo(file_key, sensor_name, sensor_info))

		// 计算分组图片列表
		group_files := Util.Group_files_by_filename(sensor_sub_files, sensor_info.Compress_interval)
		for group_key := range group_files {
			sub_files := group_files[group_key]
			file_key := fmt.Sprintf("/%s/%s/%s/%s/%d.zip", trip_date, trip_uuid, sensor_name, sensor_info.Path, group_key)
			uuid := Util.Create_uuid()

			// 任务列表
			items[uuid] = UploadItem.NewUploadItem(uuid, file_key, sub_files, sensor_info.Compress, sensor_info.ConvertWebp, false)
			// 状态列表
			stats[uuid] = false
		}
	}

	// 设置主行程信息
	trip_info.MainTrip.Mileage = trip_mileage
	trip_info.MainTrip.DataSize = trip_datasize
	trip_info.MainTrip.ImageCount = trip_imagecount
	trip_info.MainTrip.TripDate = trip_date

	// 保存任务元素列表文件
	if err := UploadItem.FlushUploadItems(&items, task_items_json_file); err != nil {
		log.Printf("Flush_to_Json Error [%s]", err)
		return err
	}

	// 保存任务列表状态文件
	if err := UploadItem.FlushUploadStats(&stats, task_stats_json_file); err != nil {
		log.Printf("FlushUploadItemsStatus Error [%s]", err)
		return err
	}

	// 构建行程信息发送结构体
	post_trip_info := &UploadItem.PostTripInfo{
		Session_uuid: trip_uuid,
		Task_info:    task_info,
		Trip_info:    trip_info,
		SensorDatas:  posts,
	}

	// 保存发送消息文件
	if err := UploadItem.FlushPostMessageInfo(post_trip_info, task_posts_json_file); err != nil {
		log.Printf("FlushPostMessageInfo Error [%s]", err)
		return err
	}

	log.Println("创建任务文件结束")

	return nil
}

var (
	help                 bool                          = false
	restart              bool                          = false
	work_dir             string                        = "/home/wanhy/下载/1128" // 工作目录路径
	task_items_json_file string                        = ""                    // 任务元素列表文件
	task_stats_json_file string                        = ""                    // 任务状态列表文件
	task_posts_json_file string                        = ""                    // 任务发送消息问津
	threadcount          int                           = 8                     // 线程数量
	trip_info            *XmlTripConfig.ConfigTripInfo = nil                   // 行程信息
	task_info            *TaskConfig.TaskInfo          = nil                   // 任务信息
	trip_mileage         float64                       = 0                     // 主行程里程
	trip_datasize        int64                         = 0                     // 数据量
	trip_imagecount      int64                         = 0                     // 图片数量
	trip_tripdate        string                        = ""                    // 行程日期

	sensors_alias2path       XmlTripConfig.Alias_To_Path
	sensors_path2alias       XmlTripConfig.Path_To_Alias
	sensors_alias2sensorinfo XmlTripConfig.Alias_To_SensorInfo
)

// 打印帮助信息
func usage() {
	log.Println(`Usage: client [-h] [-r] [-i work directory]   [-t thread count]`)
	flag.PrintDefaults()
}

// 打印参数信息
func Init_all_parameters() {
	log.SetFlags(0)
	log.Println("Starting application...")

	// log.Println("cpu:", runtime.NumCPU())
	threadcount = Util.Max(threadcount, 4)
	task_items_json_file = filepath.Join(work_dir, "./task_items.json")
	task_stats_json_file = filepath.Join(work_dir, "./task_stats.json")
	task_posts_json_file = filepath.Join(work_dir, "./post_infos.json")

	log.Println("打印参数开始...")
	log.Printf("thread count            : [%d]", threadcount)
	log.Printf("work directory          : [%s]", work_dir)
	log.Printf("task_items_json_file    : [%s]", task_items_json_file)
	log.Printf("task_stats_json_file    : [%s]", task_stats_json_file)
	log.Printf("task_posts_json_file    : [%s]", task_posts_json_file)
	log.Println("打印参数结束")
}

// 主程序入口
func main() {
	log.SetFlags(0)
	flag.BoolVar(&help, "h", false, "useage help")
	flag.BoolVar(&restart, "r", false, "restart task")
	flag.StringVar(&work_dir, "i", "", "work directory")
	flag.IntVar(&threadcount, "t", 4, "thread count")
	flag.Parse()

	img_files := []string{}
	Util.Get_All_Sub_Files("/home/wanhy/Project/git.mapbar/chezaishujucaiji/new_vehicle_collection/data_recording/record/record", &img_files)
	for _, file := range img_files {
		log.Println(file)
	}
	log.Println("file")

	if help {
		usage()
		return
	}

	// 打印已设置参数信息
	Init_all_parameters()

	// 解析服务设置文件
	server, err := ServerConfig.ParseServerConfig("./server_config.json")
	if err != nil {
		log.Printf("Config.LoadServer Error [%s]", err)
		return
	}
	// server.Print()

	// 解析行程信息
	trip_info_file := filepath.Join(work_dir, "trip_config.xml")
	trip_info, err = XmlTripConfig.ParseConfigTripInfo(trip_info_file)
	if err != nil {
		log.Printf("ParseConfigTripInfo Error [%s]", err)
		return
	}

	// 解析任务信息
	task_info_file := filepath.Join(work_dir, "task_info.json")
	task_info, err = TaskConfig.ParseTaskConfig(task_info_file)
	if err != nil {
		log.Printf("TaskInfo.ParseTaskInfo Error [%s]", err)
		return
	}

	// 设置行程日期
	trip_tripdate = task_info.Trip_date
	log.Printf("trip_tripdate   : [%s]", trip_tripdate)

	// 检查传感器文件列表
	sensors_alias2path, sensors_path2alias, sensors_alias2sensorinfo = trip_info.GetTripSensorFileList(work_dir)
	if err := Check_Sensor_Files(); err != nil {
		log.Printf("Check_Sensor_Files Error [%s]", err)
		return
	}

	// 判断任务列表文件是否存在
	if !Util.Exists(task_items_json_file) || restart {
		log.Printf("任务文件不存在，重新创建任务文件")

		// 创建任务列表文件
		if err := Create_Upload_Files_Json(); err != nil {
			log.Printf("Create_Upload_Files_Json Error [%s]", err)
			return
		}
	} else {
		log.Printf("任务文件存在，继续任务文件")
	}

	// return

	if err := UploadItem.Do_Upload_Files(server, threadcount, work_dir, task_items_json_file, task_stats_json_file, task_posts_json_file); err != nil {
		log.Printf("Do_Upload_Files Error [%v]", err)
		return
	}

	log.Println("Data Upload Finished")

}

// s3cmd ls s3://r1-iat-ntd-ob01/20201014/c3751fad-8326-4659-a658-469e221ec415/3Work/record/
// s3cmd ls s3://r1-iat-ntd-ob01/20201014/74360c84-379c-489f-9009-b91506961f86/3Work/record/
