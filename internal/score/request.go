package score

import "github.com/gin-gonic/gin"

// ScoreHistoryReq is the query parameters for the score history endpoint.
type ScoreHistoryReq struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

// RouteGuard is the interface for permission-checking middleware.
type RouteGuard interface {
	Require() gin.HandlerFunc
}
