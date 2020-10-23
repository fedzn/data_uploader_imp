package Ceph

import (
	"bytes"
	"libs/Parser/ServerConfig"
	"libs/util/Util"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type CephClient struct {
	Access_key  string //
	Secret_key  string //
	End_point   string //
	Bucket_name string //
	InsSession  *session.Session
	Uploader    *s3manager.Uploader
}

func CreateCephClient(info *ServerConfig.CephInfo) *CephClient {

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(info.Access_key, info.Secret_key, ""),
		Endpoint:         aws.String(info.End_point),
		Region:           aws.String("default"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true), //virtual-host style方式，不要修改
	}))

	return &CephClient{
		Access_key:  info.Access_key,
		Secret_key:  info.Secret_key,
		End_point:   info.End_point,
		Bucket_name: info.Bucket_name,
		InsSession:  sess,
		Uploader:    s3manager.NewUploader(sess),
	}
}

func Print_UploadOutput(file_key string, output *s3manager.UploadOutput, err error) {
	// 判断是否错误，打印错误信息
	// if err != nil {
	// 	log.Printf("s3manager.Upload Error [%v]", err)
	// } else {
	// 	log.Printf("Upload              :[%s]", file_key)
	// 	log.Printf("  Output:Location :[%s]", output.Location)
	// 	// log.Printf("	Output:VersionID  :[%s]", *output.VersionID)
	// 	log.Printf("    Output:UploadID   :[%s]", output.UploadID)
	// }
}

func (c *CephClient) Upload_Single_File(file_path string, file_key string) error {
	file, err := os.Open(file_path)
	if err != nil {
		log.Printf("os.Open Error [%v]", err)
		return err
	}
	defer file.Close()

	// 开始上传
	output, err := c.Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.Bucket_name),
		Key:    aws.String(file_key),
		Body:   file,
	})

	// 上传结束关闭文件
	file.Close()

	// 打印结果或错误提示
	Print_UploadOutput(file_key, output, err)

	return err
}

func (c *CephClient) Upload_Binary_Data(content []byte, file_key string) error {
	output, err := c.Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.Bucket_name),
		Key:    aws.String(file_key),
		Body:   bytes.NewReader(content),
	})

	// 打印结果或错误提示
	Print_UploadOutput(file_key, output, err)

	return err
}

func (c *CephClient) Upload_Multi_Files(sub_files []string, file_key string) error {
	content, err := Util.Package_files_buffer(sub_files)
	if err != nil {
		log.Printf("Util.Package_files_buffer Error [%v]", err)
		return err
	}

	return c.Upload_Binary_Data(content, file_key)
}

func (c *CephClient) Upload_Multi_Image_Files(sub_files []string, file_key string, tmp_dir string) error {

	webp_files := []string{}
	for _, sub_file := range sub_files {
		webp_file := filepath.Join(tmp_dir, filepath.Base(sub_file))
		if Util.Image_to_webp(sub_file, webp_file) {
			webp_files = append(webp_files, webp_file)
		}
	}

	defer func(sub_files []string) {
		for _, sub_file := range sub_files {
			if err := os.Remove(sub_file); err != nil {
				log.Printf("os.Remove Error [%v]", err)
			}
		}
	}(webp_files)

	return c.Upload_Multi_Files(webp_files, file_key)
}
