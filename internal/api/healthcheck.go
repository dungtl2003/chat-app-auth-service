package api

import (
	"dungtl2003/chat-app-auth-service/internal/context"
	"dungtl2003/chat-app-auth-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthCheckResponseBody struct {
	Status string            `json:"status"`
	Report map[string]string `json:"report"`
}

func HealthCheck(appCtx *context.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		report := map[string]string{}
		serverStatus := "UP"
		for _, s := range appCtx.Services {
			if s.Status() == services.ServiceReady {
				report[s.Name()] = "UP"
			} else {
				appCtx.Logger.Errorfln("Service %s is DOWN", s.Name())
				report[s.Name()] = "DOWN"
				// we don't set the overall server status to DOWN to avoid k8s
				// restarting the pod
				// serverStatus = "DOWN"

			}
		}

		responseBody := HealthCheckResponseBody{
			Status: serverStatus,
			Report: report,
		}

		appCtx.Logger.Debugfln("Response body: %v", responseBody)
		c.JSON(http.StatusOK, responseBody)
		c.Abort()
	}
}
