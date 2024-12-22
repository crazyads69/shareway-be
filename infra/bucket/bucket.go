package bucket

import (
	"context"
	"fmt"
	"mime/multipart"

	"shareway/util"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cloudinary *cloudinary.Cloudinary
}

func NewCloudinary(ctx context.Context, cfg util.Config) *CloudinaryService {
	// Cloudinary configuration
	cld, err := cloudinary.NewFromParams(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		panic(err)
	}
	// Sercure
	cld.Config.URL.Secure = true
	return &CloudinaryService{cloudinary: cld}
}

// UploadChatImage uploads an image to Cloudinary and returns the secure URL of the image file
func (c *CloudinaryService) UploadChatImage(ctx context.Context, image *multipart.FileHeader) (string, error) {
	// Open the image file
	src, err := image.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer src.Close()

	// Upload the image to Cloudinary with optimization
	uploadResult, err := c.cloudinary.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:         "chat_images",
		ResourceType:   "image",
		Transformation: "f_auto,q_auto:eco,c_limit,w_1920,h_1080",
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	return uploadResult.SecureURL, nil
}
