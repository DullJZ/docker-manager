package service

import (
	"bytes"
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// ImageService 镜像服务
type ImageService struct {
	*DockerService
}

// NewImageService 创建镜像服务
func NewImageService(dockerService *DockerService) *ImageService {
	return &ImageService{DockerService: dockerService}
}

// Delete 删除镜像
func (is *ImageService) Delete(imageName string) map[string]interface{} {
	ctx := context.Background()
	_, err := is.client.ImageRemove(ctx, imageName, types.ImageRemoveOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// GetInfo 获取镜像信息
func (is *ImageService) GetInfo(imageName string) (types.ImageInspect, error) {
	ctx := context.Background()
	info, _, err := is.client.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return types.ImageInspect{}, err
	}
	return info, nil
}

// Pull 拉取镜像
func (is *ImageService) Pull(imageName string) map[string]interface{} {
	ctx := context.Background()
	out, err := is.client.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	defer out.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	return map[string]interface{}{"status": "success", "message": buf.String()}
}

// List 列出所有镜像
func (is *ImageService) List() ([]image.Summary, error) {
	ctx := context.Background()
	images, err := is.client.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	return images, nil
}

// Search 搜索镜像
func (is *ImageService) Search(imageName string) []image.Summary {
	allImages, err := is.List()
	if err != nil {
		return []image.Summary{}
	}

	var images []image.Summary
	for _, img := range allImages {
		if len(img.RepoTags) > 0 {
			// 模糊搜索
			if fuzzy.Match(imageName, img.RepoTags[0]) {
				images = append(images, img)
			}
		}
	}
	return images
}
