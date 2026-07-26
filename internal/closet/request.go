package closet

type ListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Type     string `form:"type"`   // "skin", "cape", or empty
	Search   string `form:"search"` // keyword search on item_name
}

type AddReq struct {
	TID  uint   `json:"tid"`
	Name string `json:"name"`
}

type RenameReq struct {
	Name string `json:"name"`
}
