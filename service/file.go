package service

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
)

// FileService 文件操作服务
type FileService struct {
	*DockerService
}

// NewFileService 创建文件操作服务
func NewFileService(dockerService *DockerService) *FileService {
	return &FileService{DockerService: dockerService}
}

// CopyFromContainer 从容器拷贝文件到主机
func (fs *FileService) CopyFromContainer(containerName string, srcPath string, dstPath string) map[string]interface{} {
	ctx := context.Background()

	// 从容器获取文件内容
	reader, stat, err := fs.client.CopyFromContainer(ctx, containerName, srcPath)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	defer reader.Close()

	// 如果目标路径是目录，使用原始文件名
	if info, err := os.Stat(dstPath); err == nil && info.IsDir() {
		dstPath = filepath.Join(dstPath, stat.Name)
	}

	// 解压TAR内容到目标路径
	err = fs.extractTarToFile(reader, dstPath)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}

	return map[string]interface{}{"status": "success"}
}

// CopyToContainer 从主机拷贝文件到容器
func (fs *FileService) CopyToContainer(containerName string, srcPath string, dstPath string) map[string]interface{} {
	ctx := context.Background()

	// 检查源文件是否存在
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return map[string]interface{}{"status": "fail", "message": "source file does not exist"}
	}

	// 创建TAR内容
	tarReader, err := fs.createTarFromFile(srcPath)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	defer tarReader.Close()

	// 拷贝到容器
	err = fs.client.CopyToContainer(ctx, containerName, dstPath, tarReader, types.CopyToContainerOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}

	return map[string]interface{}{"status": "success"}
}

// extractTarToFile 从TAR流中提取文件到指定路径
func (fs *FileService) extractTarToFile(reader io.Reader, dstPath string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 只处理常规文件
		if header.Typeflag == tar.TypeReg {
			// 创建目标目录
			if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
				return err
			}

			// 创建目标文件
			outFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// 复制文件内容
			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
			break // 只提取第一个文件
		}
	}

	return nil
}

// createTarFromFile 从文件创建TAR流
func (fs *FileService) createTarFromFile(srcPath string) (io.ReadCloser, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	// 创建管道
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		defer file.Close() // 在goroutine中关闭文件

		tarWriter := tar.NewWriter(pw)
		defer tarWriter.Close()

		// 创建TAR头部
		header := &tar.Header{
			Name: filepath.Base(srcPath),
			Size: fileInfo.Size(),
			Mode: int64(fileInfo.Mode()),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			pw.CloseWithError(err)
			return
		}

		// 复制文件内容
		_, err := io.Copy(tarWriter, file)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
	}()

	return pr, nil
}

// CopyToContainerWithBinary 通过二进制数据流拷贝到容器
func (fs *FileService) CopyToContainerWithBinary(containerName string, fileName string, dstPath string, data []byte) map[string]interface{} {
	ctx := context.Background()

	// 创建TAR内容从二进制数据
	tarReader, err := fs.createTarFromBinary(fileName, data)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	defer tarReader.Close()

	// 拷贝到容器
	err = fs.client.CopyToContainer(ctx, containerName, dstPath, tarReader, types.CopyToContainerOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}

	return map[string]interface{}{"status": "success"}
}

// CopyFromContainerWithBinary 从容器读取文件到二进制数据流
func (fs *FileService) CopyFromContainerWithBinary(containerName string, srcPath string) ([]byte, map[string]interface{}) {
	ctx := context.Background()

	// 从容器获取文件内容
	reader, _, err := fs.client.CopyFromContainer(ctx, containerName, srcPath)
	if err != nil {
		return nil, map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	defer reader.Close()

	// 解压TAR内容到二进制数据
	data, err := fs.extractTarToBinary(reader)
	if err != nil {
		return nil, map[string]interface{}{"status": "fail", "message": err.Error()}
	}

	return data, map[string]interface{}{"status": "success"}
}

// createTarFromBinary 从二进制数据创建TAR流
func (fs *FileService) createTarFromBinary(fileName string, data []byte) (io.ReadCloser, error) {
	// 创建管道
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		tarWriter := tar.NewWriter(pw)
		defer tarWriter.Close()

		// 创建TAR头部
		header := &tar.Header{
			Name: fileName,
			Size: int64(len(data)),
			Mode: 0644,
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			pw.CloseWithError(err)
			return
		}

		// 写入二进制数据
		_, err := tarWriter.Write(data)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
	}()

	return pr, nil
}

// extractTarToBinary 从TAR流中提取文件到二进制数据
func (fs *FileService) extractTarToBinary(reader io.Reader) ([]byte, error) {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// 只处理常规文件
		if header.Typeflag == tar.TypeReg {
			// 读取文件内容到内存
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}
			return data, nil // 返回第一个文件的内容
		}
	}

	return nil, io.EOF // 没有找到文件
}
