/*
 * Copyright (c) 2025.  Kamesh Sampath <kamesh.sampath@hotmail.com>
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 *
 */

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
