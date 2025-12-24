package util

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"wxcloudrun-golang/consts"
)

// 因为有 docker app 绝对路径，所以这里都要拼接下

func WriteFileSafe(path string, data []byte) error {
	dir := filepath.Dir(consts.AbsolutePath + path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(consts.AbsolutePath+path, data, 0644)
}

func BizOsStat(path string) (os.FileInfo, error) {
	return os.Stat(consts.AbsolutePath + path)
}

func BizOsMkdirAll(path string) error {
	return os.MkdirAll(consts.AbsolutePath+path, 0755)
}

func BizOsReadFile(path string) ([]byte, error) {
	return os.ReadFile(consts.AbsolutePath + path)
}

func SaveUploadedFileManual(file *multipart.FileHeader, dst string) error {
	// 1. 创建目录
	if err := BizOsMkdirAll(filepath.Dir(dst)); err != nil {
		return err
	}

	// 2. 打开上传文件
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 3. 创建目标文件
	dstFile, err := os.OpenFile(consts.AbsolutePath+dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 4. 拷贝内容
	_, err = io.Copy(dstFile, src)
	return err
}
