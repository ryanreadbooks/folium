package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanreadbooks/folium/internal/segment/idgen"
)

var (
	httpServer *http.Server
	eng        *gin.Engine
)

func InitHttp() {
	idgen.Init()
	initRoute()

	go http.ListenAndServe(":9527", eng)
}

func initRoute() {
	eng = gin.Default()

	// /api/v1/:key?step=xxx
	eng.GET("/api/v1/:key", nextForKey)
}

type Result struct {
	Id  uint64 `json:"id,omitempty"`
	Msg string `json:"msg,omitempty"`
}

func nextForKey(c *gin.Context) {
	key := c.Param("key")
	step := c.Query("step")
	if len(step) != 0 {
		ss, err := strconv.Atoi(step)
		if err == nil {
			// getIdWithStep
			idgen.GetNextWithStep(c, key, uint32(ss))
			return
		}
	}

	id, err := idgen.GetNext(c, key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, &Result{
			Msg: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &Result{
		Id: id,
	})
}
