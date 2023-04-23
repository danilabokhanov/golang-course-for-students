package httpgin

import (
	"errors"
	"github.com/gin-gonic/gin"
	"homework9/internal/app"
	"net/http"
	"strconv"
	"time"
)

func createAd(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {

		var reqBody createAdRequest
		err := c.ShouldBindJSON(&reqBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		ad, e := a.CreateAd(c, reqBody.Title, reqBody.Text, reqBody.UserID)

		if e != nil {
			if errors.Is(e, app.ErrWrongFormat) {
				c.JSON(http.StatusBadRequest, ErrorResponse(e))
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse(e))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		c.JSON(http.StatusOK, AdSuccessResponse(&ad))
	}
}

func changeAdStatus(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody changeAdStatusRequest
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		strAdID := c.Param("ad_id")
		adID, err := strconv.Atoi(strAdID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}
		_, err = a.FindAd(c, int64(adID))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		ad, e := a.ChangeAdStatus(c, int64(adID), reqBody.UserID, reqBody.Published)
		if e != nil {
			if errors.Is(e, app.ErrWrongFormat) {
				c.JSON(http.StatusBadRequest, ErrorResponse(e))
				return
			}
			if errors.Is(e, app.ErrNoAccess) {
				c.JSON(http.StatusForbidden, ErrorResponse(e))
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse(e))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		c.JSON(http.StatusOK, AdSuccessResponse(&ad))
	}
}

// Метод для обновления текста(Text) или заголовка(Title) объявления
func updateAd(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody updateAdRequest
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
		}

		strAdID := c.Param("ad_id")
		adID, err := strconv.Atoi(strAdID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}
		_, err = a.FindAd(c, int64(adID))
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		ad, e := a.UpdateAd(c, int64(adID), reqBody.UserID, reqBody.Title, reqBody.Text)
		if e != nil {
			if errors.Is(e, app.ErrWrongFormat) {
				c.JSON(http.StatusBadRequest, ErrorResponse(e))
				return
			}
			if errors.Is(e, app.ErrNoAccess) {
				c.JSON(http.StatusForbidden, ErrorResponse(e))
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse(e))
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		c.JSON(http.StatusOK, AdSuccessResponse(&ad))
	}
}

func listAds(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		f, err := a.GetNewFilter(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		filter, err := f.BasicConfig(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		strAuthorID := c.Query("author_id")
		if strAuthorID != "" {
			authorID, err := strconv.Atoi(strAuthorID)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
			filter, err = filter.SetAuthor(c, int64(authorID))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
		}

		strPublishedOnly := c.Query("published_only")
		if strPublishedOnly != "" {
			publishedOnly, err := strconv.ParseBool(strPublishedOnly)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
			filter, err = filter.SetStatus(c, publishedOnly)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
		}

		strLTime := c.Query("l_time")
		if strLTime != "" {
			seconds, err := strconv.ParseInt(strLTime, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
			lTime := time.UnixMicro(seconds).UTC()
			filter, err = filter.SetLTime(c, lTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
		}

		strRTime := c.Query("r_time")
		if strRTime != "" {
			seconds, err := strconv.ParseInt(strRTime, 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
			rTime := time.UnixMicro(seconds).UTC()
			filter, err = filter.SetRTime(c, rTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
		}

		pattern, err := filter.GetPattern(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		ads, e := a.GetAllAdsByTemplate(c, pattern)
		if e != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		c.JSON(http.StatusOK, AdSuccessResponseList(&ads))
	}
}

func getAdByID(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		strAdID := c.Param("ad_id")
		adID, err := strconv.Atoi(strAdID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		ad, err := a.FindAd(c, int64(adID))
		if err != nil {
			if errors.Is(err, app.ErrWrongFormat) {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		c.JSON(http.StatusOK, AdSuccessResponse(&ad))
	}
}

func deleteAd(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		strAdID := c.Param("ad_id")
		adID, err := strconv.Atoi(strAdID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		var reqBody deleteAdRequest
		err = c.ShouldBindJSON(&reqBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		ad, e := a.DeleteAd(c, int64(adID), reqBody.UserID)
		if e != nil {
			if errors.Is(e, app.ErrWrongFormat) {
				c.JSON(http.StatusBadRequest, ErrorResponse(e))
				return
			}
			if errors.Is(e, app.ErrNoAccess) {
				c.JSON(http.StatusForbidden, ErrorResponse(e))
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse(e))
			return
		}

		c.JSON(http.StatusOK, AdSuccessResponse(&ad))
	}
}

func createUser(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {

		var reqBody universalUser
		err := c.ShouldBindJSON(&reqBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		u, e := a.CreateUserByID(c, reqBody.Nickname, reqBody.Email, reqBody.ID)

		if e != nil {
			if errors.Is(e, app.ErrWrongFormat) {
				c.JSON(http.StatusBadRequest, ErrorResponse(e))
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse(e))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		c.JSON(http.StatusOK, UserSuccessResponse(&u))
	}
}

func deleteUserByID(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		strUserID := c.Param("user_id")
		userID, err := strconv.Atoi(strUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		u, err := a.DeleteUserByID(c, int64(userID))
		if err != nil {
			if errors.Is(err, app.ErrWrongFormat) {
				c.JSON(http.StatusBadRequest, ErrorResponse(err))
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}

		c.JSON(http.StatusOK, UserSuccessResponse(&u))
	}
}

func changeUserInfo(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody changeUserStatusRequest
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		strUserID := c.Param("user_id")
		userID, err := strconv.Atoi(strUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}
		_, isFound, err := a.FindUser(c, int64(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		if !isFound {
			c.Status(http.StatusBadRequest)
			return
		}

		u, e := a.ChangeUserInfo(c, int64(userID), reqBody.Nickname, reqBody.Email)
		if e != nil {
			if errors.Is(e, app.ErrWrongFormat) {
				c.JSON(http.StatusBadRequest, ErrorResponse(e))
				return
			}
			if errors.Is(e, app.ErrNoAccess) {
				c.JSON(http.StatusForbidden, ErrorResponse(e))
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse(e))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		c.JSON(http.StatusOK, UserSuccessResponse(&u))
	}
}

func getAdsByTitle(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {

		title := c.Query("title")
		ads, err := a.GetAdsByTitle(c, title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		c.JSON(http.StatusOK, AdSuccessResponseList(&ads))
	}
}

func getUserByID(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		strUserID := c.Param("user_id")
		userID, err := strconv.Atoi(strUserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse(err))
			return
		}

		u, isFound, err := a.FindUser(c, int64(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse(err))
			return
		}
		if !isFound {
			c.Status(http.StatusBadRequest)
			return
		}

		c.JSON(http.StatusOK, UserSuccessResponse(&u))
	}
}
