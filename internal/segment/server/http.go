package server

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ryanreadbooks/folium/internal/segment/idgen"
)

var (
	serverHttp *http.Server
	eng        *gin.Engine
)

func init() {
	idgen.Init()
}

func CloseServer() {
	CloseHttp()
	CloseGrpc()
	idgen.Close()
}

func InitHttp() {
	initRoute()

	go func() {
		if err := http.ListenAndServe(":9527", eng); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
}

func CloseHttp() {
	if serverHttp != nil {
		serverHttp.Close()
	}
}

func initRoute() {
	eng = gin.Default()

	// /api/v1/next/:key?step=xxx
	eng.GET("/api/v1/next/:key", nextForKey)
}

type Result struct {
	Id  uint64 `json:"id,omitempty"`
	Msg string `json:"msg,omitempty"`
}

func nextForKey(c *gin.Context) {
	key := c.Param("key")
	step := c.Query("step")
	stepNum, _ := strconv.Atoi(step)
	id, err := idgen.GetNext(c, key, idgen.WithStep(uint32(stepNum)))
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
