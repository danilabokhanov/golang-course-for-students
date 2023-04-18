package httpgin

import (
	"github.com/gin-gonic/gin"
	"homework8/internal/app"
)

func AppRouter(r *gin.RouterGroup, a app.App) {
	r.POST("/ads", createAd(a))
	r.PUT("/ads/:ad_id/status", changeAdStatus(a))
	r.PUT("/ads/:ad_id", updateAd(a))
	r.GET("/ads", listAds(a))
	r.GET("/ads/by_title", getAdsByTitle(a))
	r.GET("/ads/:ad_id", getAdByID(a))
	r.POST("/users", createUser(a))
	r.PUT("/users/:user_id", changeUserInfo(a))
	r.GET("/users/:user_id", getUserByID(a))
}
