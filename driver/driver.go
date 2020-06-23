// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a implementation of a ProtocolDriver interface.
//
package driver

import (
	"fmt"
	"github.com/spf13/cast"
	"strconv"
	"sync"
	"time"

	COAP "github.com/dustin/go-coap"
	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var once sync.Once
var driver *Driver
var CurrentMessageID = 12345

type Driver struct {
	lc            logger.LoggingClient
	asyncCh       chan<- *dsModels.AsyncValues
}

func NewProtocolDriver() dsModels.ProtocolDriver {
	once.Do(func() {
		driver = new(Driver)
	})
	return driver
}

func (d *Driver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.lc.Info(fmt.Sprintf("Driver.DisconnectDevice: device-coap-go driver is disconnecting to %s", deviceName))
	return nil
}

func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues) error {
	d.lc = lc
	d.asyncCh = asyncCh
	return nil
}

func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	var responses = make([]*dsModels.CommandValue, len(reqs))

	var connectionInfo = protocols["coap"]
	addr := connectionInfo["Address"]

    client, err := COAP.Dial("udp",addr)

    if err != nil{
    	return responses,nil
	}

	for i, req := range reqs{
		res, err := d.handleReadCommandRequest(client, req)
		if err != nil{
			d.lc.Info(fmt.Sprintf("Handle read commands failed: %v", err))
			return responses,nil
		}

		responses[i] = res
	}

	return responses, nil
}

func (d *Driver) handleReadCommandRequest(deviceClient *COAP.Conn, req dsModels.CommandRequest) (*dsModels.CommandValue, error) {
	var result = &dsModels.CommandValue{}
	var err error

	requ := COAP.Message{
		Type: COAP.Confirmable,
		Code: COAP.GET,
		MessageID: GenerateMessageID(),
	}

	if req.DeviceResourceName == "rand" {
		path := "/rand"
		requ.SetPathString(path)
		resp, err := deviceClient.Send(requ)
		if err != nil{
			d.lc.Error(fmt.Sprintf("Driver.handleReadCommands: Read failed: %s", err))
		}
		s := resp.Payload
		num := string(s)
		reading,_ := strconv.Atoi(num)
		result, err = newResult(req, reading)
		if err != nil {
			return result, err
		} else {
			d.lc.Info(fmt.Sprintf("Get command finished: %v", result))
		}
	}else {
		path := "/ping"
		requ.SetPathString(path)
		resp, err := deviceClient.Send(requ)
		if err != nil{
			d.lc.Error(fmt.Sprintf("Driver.handleReadCommands: Read failed: %s", err))
		}
		reading := resp.Payload
		result, err = newResult(req, reading)
		if err != nil {
			return result, err
		} else {
			d.lc.Info(fmt.Sprintf("Get command finished: %v", result))
		}
	}
	return result, err
}

func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {

	var connectionInfo = protocols["coap"]
	addr := connectionInfo["Address"]

	client, err := COAP.Dial("udp",addr)

	if err != nil{
		return err
	}

	for _, req := range reqs {
		for _, param := range params{
			err := d.handleWriteCommandRequest(client, req, param)
			if err != nil {
				d.lc.Info(fmt.Sprintf("Handle write commands failed: %v", err))
				return err
			}
		}
	}

	return nil
}

func (d *Driver) handleWriteCommandRequest(deviceClient *COAP.Conn, req dsModels.CommandRequest, param *dsModels.CommandValue) error {

	var err error

	value, err := param.Int64Value()
	num := strconv.Itoa(int(value))

	request := COAP.Message{
		Type: COAP.NonConfirmable,
		Code: COAP.PUT,
		MessageID: GenerateMessageID(),
		Payload: []byte(num),
	}

	request.SetPathString("/rand")
	_, err = deviceClient.Send(request)

	if err != nil {
		d.lc.Error(fmt.Sprintf("Driver.handleWriteCommands: Write value %v failed: %s", value, err))
		return err
	}

	return nil
}

func (d *Driver) Stop(force bool) error {
	d.lc.Info("Driver.Stop: device-coap-go driver is stopping...")
	return nil
}

func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.lc.Debug(fmt.Sprintf("a new Device is added: %s", deviceName))
	return nil
}

func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.lc.Debug(fmt.Sprintf("Device %s is updated", deviceName))
	return nil
}

func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.lc.Debug(fmt.Sprintf("Device %s is removed", deviceName))
	return nil
}

func newResult(req dsModels.CommandRequest, reading interface{}) (*dsModels.CommandValue, error) {
	var result = &dsModels.CommandValue{}
	var err error
	var resTime = time.Now().UnixNano() / int64(time.Millisecond)
	castError := "fail to parse %v reading, %v"

	if !checkValueInRange(req.Type, reading) {
		err = fmt.Errorf("parse reading fail. Reading %v is out of the value type(%v)'s range", reading, req.Type)
		driver.lc.Error(err.Error())
		return result, err
	}

	switch req.Type {
	case dsModels.Bool:
		val, err := cast.ToBoolE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewBoolValue(req.DeviceResourceName, resTime, val)
	case dsModels.String:
		val, err := cast.ToStringE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result = dsModels.NewStringValue(req.DeviceResourceName, resTime, val)
	case dsModels.Uint8:
		val, err := cast.ToUint8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewUint8Value(req.DeviceResourceName, resTime, val)
	case dsModels.Uint16:
		val, err := cast.ToUint16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewUint16Value(req.DeviceResourceName, resTime, val)
	case dsModels.Uint32:
		val, err := cast.ToUint32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewUint32Value(req.DeviceResourceName, resTime, val)
	case dsModels.Uint64:
		val, err := cast.ToUint64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewUint64Value(req.DeviceResourceName, resTime, val)
	case dsModels.Int8:
		val, err := cast.ToInt8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewInt8Value(req.DeviceResourceName, resTime, val)
	case dsModels.Int16:
		val, err := cast.ToInt16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewInt16Value(req.DeviceResourceName, resTime, val)
	case dsModels.Int32:
		val, err := cast.ToInt32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewInt32Value(req.DeviceResourceName, resTime, val)
	case dsModels.Int64:
		val, err := cast.ToInt64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewInt64Value(req.DeviceResourceName, resTime, val)
	case dsModels.Float32:
		val, err := cast.ToFloat32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewFloat32Value(req.DeviceResourceName, resTime, val)
	case dsModels.Float64:
		val, err := cast.ToFloat64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
		}
		result, err = dsModels.NewFloat64Value(req.DeviceResourceName, resTime, val)
	default:
		err = fmt.Errorf("return result fail, none supported value type: %v", req.Type)
	}
	return result, err
}

func GenerateMessageID() uint16 {
	if CurrentMessageID != 65535 {
		CurrentMessageID++
	} else {
		CurrentMessageID = 10000
	}
	return uint16(CurrentMessageID)
}