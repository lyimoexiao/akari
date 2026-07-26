package closet

import "errors"

var (
	ErrClosetItemNotFound  = errors.New("衣橱物品不存在")
	ErrAlreadyInCloset     = errors.New("已在衣橱中")
	ErrNotEnoughScore      = errors.New("积分不足")
	ErrNotTextureUploader  = errors.New("只有材质上传者才能重命名")
)
