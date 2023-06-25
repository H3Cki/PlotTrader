package main

import (
	"github.com/H3Cki/Plotor/controllers"
	"github.com/H3Cki/Plotor/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Managing plot orders
	r.POST("/plotorder", controllers.CreatePlotOrder())
	r.GET("/plotorder", controllers.GetPlotOrder())
	r.DELETE("/plotorder", controllers.CancelPlotOrder())
	//r.POST("/attach", controllers.Attach())

	// Managing Sessions
	r.POST("/session", controllers.CreateSession())
	r.GET("/session", controllers.GetSessions())
	r.DELETE("/session", controllers.DeleteSession())

	if err := r.Run(); err != nil {
		logger.Error(err)
	} // listen and serve on 0.0.0.0:8080
}
