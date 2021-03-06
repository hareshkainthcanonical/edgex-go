//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"fmt"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type DeviceServiceController struct {
	reader io.DeviceServiceReader
	dic    *di.Container
}

// NewDeviceServiceController creates and initializes an DeviceServiceController
func NewDeviceServiceController(dic *di.Container) *DeviceServiceController {
	return &DeviceServiceController{
		reader: io.NewDeviceServiceRequestReader(),
		dic:    dic,
	}
}

func (dc *DeviceServiceController) AddDeviceService(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	addDeviceServiceDTOs, err := dc.reader.ReadAddDeviceServiceRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		// Encode and send the resp body as JSON format
		pkg.Encode(errResponses, w, lc)
		return
	}
	deviceServices := requestDTO.AddDeviceServiceReqToDeviceServiceModels(addDeviceServiceDTOs)

	var addResponses []interface{}
	for i, d := range deviceServices {
		newId, err := application.AddDeviceService(d, ctx, dc.dic)
		var addDeviceServiceResponse interface{}
		// get the requestID from addDeviceServiceDTOs
		reqId := addDeviceServiceDTOs[i].RequestId

		if err == nil {
			addDeviceServiceResponse = commonDTO.NewBaseWithIdResponse(
				reqId,
				fmt.Sprintf("Add device service %s successfully", d.Name),
				http.StatusCreated,
				newId)
		} else {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			addDeviceServiceResponse = commonDTO.NewBaseResponse(
				reqId,
				err.Error(),
				err.Code())
		}
		addResponses = append(addResponses, addDeviceServiceResponse)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	// Encode and send the resp body as JSON format
	pkg.Encode(addResponses, w, lc)
}

func (dc *DeviceServiceController) GetDeviceServiceByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	deviceService, err := application.GetDeviceServiceByName(name, ctx, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = responseDTO.NewDeviceServiceResponseNoMessage("", http.StatusOK, deviceService)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceServiceController) PatchDeviceService(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	updateDeviceServiceDTOs, err := dc.reader.ReadUpdateDeviceServiceRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(errResponses, w, lc)
		return
	}

	var updateResponses []interface{}
	for _, dto := range updateDeviceServiceDTOs {
		var response interface{}
		reqId := dto.RequestId
		err := application.PatchDeviceService(dto.Service, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			response = commonDTO.NewBaseResponse(
				reqId,
				"",
				http.StatusOK)
		}
		updateResponses = append(updateResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(updateResponses, w, lc)
}

func (dc *DeviceServiceController) DeleteDeviceServiceById(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	id := vars[contractsV2.Id]

	var response interface{}
	var statusCode int

	err := application.DeleteDeviceServiceById(id, ctx, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse(
			"",
			"Delete device service successfully",
			http.StatusOK)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceServiceController) DeleteDeviceServiceByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	err := application.DeleteDeviceServiceByName(name, ctx, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse(
			"",
			"Delete device service successfully",
			http.StatusOK)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceServiceController) GetAllDeviceServices(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var labels = []string{}
	var response interface{}
	var statusCode int

	// parse URL query string
	offset, err := utils.ParseQueryStringToInt(r, contractsV2.Offset, contractsV2.DefaultOffset)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	}

	limit, err := utils.ParseQueryStringToInt(r, contractsV2.Limit, contractsV2.DefaultLimit)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	}

	labels = utils.ParseQueryStringToStrings(r, contractsV2.Labels, contractsV2.CommaSeparator)

	deviceServices, err := application.GetDeviceServices(offset, limit, labels, ctx, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = responseDTO.NewMultiDeviceServicesResponse("", "", http.StatusOK, deviceServices)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}
