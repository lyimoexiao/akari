package skinlib

type ListReq struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Type     string `form:"type"`                 // "skin", "cape", or empty
	Search   string `form:"search"`               // keyword search
	Order    string `form:"order,default=latest"` // "latest" or "likes"
	Uploader uint   `form:"uploader"`             // filter by uploader (manage only)
}

type UpdateReq struct {
	Name   string `json:"name"`
	Public bool   `json:"public"`
}
