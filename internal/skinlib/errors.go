package skinlib

import "errors"

var (
	ErrTextureNotFound     = errors.New("纹理不存在")
	ErrForbidden           = errors.New("无权操作此纹理")
	ErrInvalidType         = errors.New("纹理类型无效，仅支持 steve/alex/cape")
	ErrAlreadyUploaded     = errors.New("您已上传此材质")
)
