package routes

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// Live checks for the liveliness of the API endpoints
func (e *EndpointConfig) Live(c echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}

// Ready checks for the readiness of the API dependency, especially KafkaProducer
func (e *EndpointConfig) Ready(c echo.Context) error {
	if err := e.KafkaProducer.Start(); err == nil {
		return nil
	}

	return c.JSON(http.StatusNotFound, "YDAER")
}
