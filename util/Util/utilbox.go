package Util

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func Is_exist_file(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

// 获取文件夹内的所有文件
func Get_All_Sub_Files(work_dir string, all_files *[]string) {
	dirs, err := ioutil.ReadDir(work_dir)
	if err != nil {
		log.Printf("ioutil.ReadDir Error %s", err)
		return
	}
	for _, fi := range dirs {
		if fi.IsDir() {
			sub_dir := filepath.Join(work_dir, fi.Name())
			Get_All_Sub_Files(sub_dir, all_files)
		} else {
			file_path := filepath.Join(work_dir, fi.Name())
			// log.Println(file_path)
			*all_files = append(*all_files, file_path)
		}
	}
}

func Get_all_files(work_dir string, ignore_null_file bool) []string {
	all_files := []string{}
	files, _ := ioutil.ReadDir(work_dir)
	for _, file := range files {
		file_name := file.Name()
		file_size := file.Size()
		file_path := filepath.Join(work_dir, file_name)

		if ignore_null_file && (file_size < 1) {
			continue
		}

		all_files = append(all_files, file_path)
	}

	return all_files
}

func Group_files_by_filename(filepaths []string, interval uint) map[int]([]string) {

	group_imgs := make(map[int]([]string))
	for _, file_path := range filepaths {
		file_name := filepath.Base(file_path)

		file_ext := filepath.Ext(file_name)

		v := strings.Trim(file_name, file_ext)
		file_time, err := strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}

		if file_time < 1 {
			continue
		}

		group_key := int(file_time / float64(interval))

		if _, ok := group_imgs[group_key]; ok {
			lst := group_imgs[group_key]
			lst = append(lst, file_path)
			group_imgs[group_key] = lst
		} else {
			lst := []string{file_path}
			group_imgs[group_key] = lst
		}

		// log.Println(file_name, file_ext, file_time, group_key)
	}

	// log.Println(group_imgs)

	return group_imgs
}

func Timestamp_to_date(t int64) string {
	tt := time.Unix(t/1000, 0)
	// start := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	// tt := start.Add(time.Milliseconds * t)
	// return tt.Format("19700708")
	// log.Println(tt)
	return fmt.Sprintf("%d%02d%02d", tt.Year(), tt.Month(), tt.Day())
}

func Create_uuid() string {
	return uuid.NewV4().String()
}

// 判断文件是否有效，现在仅判断jpeg格式的
func Check_valid_image(img_file string) bool {
	f1, err := os.Open(img_file)
	defer f1.Close()
	if err != nil {
		log.Printf("Open [%s] error [%s]", img_file, err)
		return false
	}

	if _, err := jpeg.Decode(f1); err != nil {
		log.Printf("jpeg.Decode [%s] error [%s]", img_file, err)
		return false
	}

	f1.Close()

	return true
}

// 调用cwebp将图片转为webp格式
func Image_to_webp(img_file string, webp_file string) bool {
	// line = r"cwebp -quiet -resize 640 360 {} -o {}".format(jpg_file, webp_file)
	// # os.popen(line, 'r')
	// os.system(line)
	// cmd_line := fmt.Sprinf("cwebp -quiet -resize 640 360 %s -o %s".format(jpg_file, webp_file)
	cmd := exec.Command("cwebp", "-quiet", "-resize", "640", "360", img_file, "-o", webp_file)

	if err := cmd.Start(); err != nil {
		log.Printf("image_to_webp Start [%s] error [%s]", img_file, err)
		return false
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("image_to_webp Wait [%s] error [%s]", img_file, err)
		return false
	}

	return true
}

func Package_files(zip_file string, img_files []string) error {
	buf, err := Package_files_buffer(img_files)
	if err != nil {
		log.Printf("Package_files_buffer Error [%v]", err)
		return err
	}

	// 将压缩文档内容写入文件
	if err := ioutil.WriteFile(zip_file, buf, 0666); err != nil {
		log.Printf("ioutil.WriteFile Error [%v]", err)
		return err
	}

	return nil
}

func Package_files_buffer(img_files []string) ([]byte, error) {
	// 创建一个缓冲区用来保存压缩文件内容
	buf := new(bytes.Buffer)

	// 创建一个压缩文档
	w := zip.NewWriter(buf)

	for _, img_file := range img_files {

		finfo, err := os.Stat(img_file)
		if err != nil {
			log.Printf("os.Stat Error [%v]", err)
			continue
		}

		f, err := w.Create(finfo.Name())
		if err != nil {
			log.Printf("zip Create Error [%v]", err)
			continue
		}

		content, err := ioutil.ReadFile(img_file)
		if err != nil {
			log.Printf("ioutil.ReadFile Error [%v]", err)
			continue
		}

		if _, err = f.Write(content); err != nil {
			log.Printf("zip Write Error [%v]", err)
			continue
		}
	}

	// 关闭压缩文档
	if err := w.Close(); err != nil {
		log.Printf("zip Write Error [%v]", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

//转化为弧度(rad)
func rad(d float64) (r float64) {
	r = d * math.Pi / 180.0
	return
}

// https://www.cnblogs.com/aoldman/p/4241117.html
// Haversine公式实现

func ToRadians(degree float64) float64 {

	return degree * math.Pi / 180.0
}

// 返回值单位米
func Distance_Haversine(longitude1, latitude1, longitude2, latitude2 float64) float64 {

	// R is the radius of the earth in kilometers (mean radius = 6,371km);
	var R float64 = 6371

	var deltaLatitude float64 = ToRadians(latitude2 - latitude1)

	var deltaLongitude float64 = ToRadians(longitude2 - longitude1)

	latitude1 = ToRadians(latitude1)

	latitude2 = ToRadians(latitude2)

	var a float64 = math.Sin(deltaLatitude/2.0)*math.Sin(deltaLatitude/2.0) + math.Cos(latitude1)*math.Cos(latitude2)*math.Sin(deltaLongitude/2.0)*math.Sin(deltaLongitude/2.0)

	var c float64 = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	var d float64 = R * c

	return d * 1000
}

// http://www.movable-type.co.uk/scripts/latlong.html
func Distance_Great_circle(lon1, lat1, lon2, lat2 float64) float64 {
	var R float64 = 6371e3                         // metres
	var latitude1 float64 = lat1 * math.Pi / 180.0 // φ, λ in radians
	var latitude2 float64 = lat2 * math.Pi / 180.0
	var deltaLatitude float64 = (lat2 - lat1) * math.Pi / 180.0
	var deltaLongitude float64 = (lon2 - lon1) * math.Pi / 180.0

	var a float64 = math.Sin(deltaLatitude/2.0)*math.Sin(deltaLatitude/2.0) +
		math.Cos(latitude1)*math.Cos(latitude2)*
			math.Sin(deltaLongitude/2.0)*math.Sin(deltaLongitude/2.0)
	var c float64 = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	var d = R * c // in metres

	return d
}

func latitude_longitude_distance_error(lon1, lat1, lon2, lat2 float64) float64 {
	//赤道半径(单位m)
	const EARTH_RADIUS = 6378137
	rad_lat1 := rad(lat1)
	rad_lon1 := rad(lon1)
	rad_lat2 := rad(lat2)
	rad_lon2 := rad(lon2)
	if rad_lat1 < 0 {
		rad_lat1 = math.Pi/2 + math.Abs(rad_lat1)
	}
	if rad_lat1 > 0 {
		rad_lat1 = math.Pi/2 - math.Abs(rad_lat1)
	}
	if rad_lon1 < 0 {
		rad_lon1 = math.Pi*2 - math.Abs(rad_lon1)
	}
	if rad_lat2 < 0 {
		rad_lat2 = math.Pi/2 + math.Abs(rad_lat2)
	}
	if rad_lat2 > 0 {
		rad_lat2 = math.Pi/2 - math.Abs(rad_lat2)
	}
	if rad_lon2 < 0 {
		rad_lon2 = math.Pi*2 - math.Abs(rad_lon2)
	}
	x1 := EARTH_RADIUS * math.Cos(rad_lon1) * math.Sin(rad_lat1)
	y1 := EARTH_RADIUS * math.Sin(rad_lon1) * math.Sin(rad_lat1)
	z1 := EARTH_RADIUS * math.Cos(rad_lat1)

	x2 := EARTH_RADIUS * math.Cos(rad_lon2) * math.Sin(rad_lat2)
	y2 := EARTH_RADIUS * math.Sin(rad_lon2) * math.Sin(rad_lat2)
	z2 := EARTH_RADIUS * math.Cos(rad_lat2)
	d := math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2) + (z1-z2)*(z1-z2))
	theta := math.Acos((EARTH_RADIUS*EARTH_RADIUS + EARTH_RADIUS*EARTH_RADIUS - d*d) / (2 * EARTH_RADIUS * EARTH_RADIUS))
	distance := theta * EARTH_RADIUS
	return distance
}

func Calc_cpt_distance(cpt_file string, t0 int64, t1 int64) (float64, error) {
	f, err := os.Open(cpt_file)
	if err != nil {
		log.Printf("os.Open Error [%v]", err)
		return 0, err
	}

	// log.Println("Calc_cpt_distance ", cpt_file, t0, t1)

	distance := 0.0
	is_first := true
	last_x := 0.0
	last_y := 0.0

	//建立缓冲区，把文件内容放到缓冲区中
	buf := bufio.NewReader(f)
	for {
		//遇到\n结束读取
		b, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("buf.ReadBytes Error [%v]", err)
				return 0, err
			}
		}

		// fmt.Println(string(b))

		fs := strings.Split(string(b), ",")
		if len(fs) < 5 {
			continue
		}

		t, errt := strconv.ParseInt(fs[0], 10, 64)
		x, errx := strconv.ParseFloat(fs[1], 64)
		y, erry := strconv.ParseFloat(fs[2], 64)
		// z, errz := strconv.ParseFloat(fs[3], 64)

		if errt != nil || errx != nil || erry != nil {
			log.Printf("strconv.ParseInt [%v] [%v] [%v]", errt, errx, erry)
			continue
		}

		if t < t0 {
			continue
		}

		if t > t1 {
			break
		}

		if is_first {
			last_x = x
			last_y = y
			is_first = false
		} else {
			distance += Distance_Haversine(last_x, last_y, x, y)
			// log.Println(distance)
			// distance += latitude_longitude_distance(last_x, last_y, x, y)
			// distance += Distance_Great_circle(last_x, last_y, x, y)
			last_x = x
			last_y = y
		}
	}

	return distance, nil
}

func MakeDir(dir_path string) error {
	// err := os.Mkdir(dir_path, os.ModePerm)
	err := os.MkdirAll(dir_path, os.ModePerm)
	if err != nil {
		log.Printf("os.Mkdir [%s]", err)
		return err
	}

	return nil
}

func Calc_Folder_Capacity(work_dir string) (int64, error) {
	var size int64 = 0
	err := filepath.Walk(work_dir, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func GetFileSize(path string) int64 {
	if !Exists(path) {
		return 0
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

//exists Whether the path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// 获取指定数组的随机索引值
func GetRandArrayIndex(arrayCount int) int {
	rand.Seed(time.Now().Unix())
	index := rand.Intn(arrayCount)

	return index
}

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func SendRequest(is_post bool, url string, body_text string) ([]byte, error) {
	RequestType := "POST"
	if !is_post {
		RequestType = "GET"
	}

	request, err := http.NewRequest(RequestType, url, strings.NewReader(body_text))
	if err != nil {
		log.Printf("http.NewRequest,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Connection", "Keep-Alive")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("http.Do failed,[err=%s][url=%s]", err, url)
		return []byte(""), err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("http.Do failed,[err=%s][url=%s]", err, url)
	}
	return b, err
}

func getKeys(m map[int]([]string)) []int {
	// keys := make([]int, len(m))
	keys := []int{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// 是否存在 大小 文件数量 空文件数量
func GetSensorFileInfo(file_path string) (bool, int64, int64, int64) {
	file_info, err := os.Stat(file_path)
	is_exist := (err == nil || os.IsExist(err))
	var file_size int64 = 0
	var file_count int64 = 0
	var null_file_count int64 = 0

	if !is_exist {
		return is_exist, file_size, file_count, null_file_count
	}

	if file_info.IsDir() {
		filepath.Walk(file_path, func(_ string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				file_size += info.Size()
				file_count += 1
				if file_size < 1 {
					null_file_count += 1
				}
			}
			return err
		})
	} else {
		file_size += file_info.Size()
		file_count += 1
		if file_size < 1 {
			null_file_count += 1
		}
	}

	return is_exist, file_size, file_count, null_file_count
}

// func Test_files_func() {
// 	work_dir := "/home/wanhy/Project/git.mapbar/chezaishujucaiji/vehicle-data-recorder/"
// 	work_dir = "/home/wanhy/下载/1128/record"
// 	work_dir = "/home/wanhy/下载/1128/rawdata-6/image"
// 	all_files := Get_all_files(work_dir, true)
// 	log.Println(all_files, len(all_files))
// }
