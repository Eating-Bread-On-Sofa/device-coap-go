// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/edgexfoundry/device-coap-go"
	"github.com/edgexfoundry/device-coap-go/driver"
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
)

const (
	serviceName string = "edgex-device-coap"
)

func main() {
	d := driver.NewProtocolDriver()
	startup.Bootstrap(serviceName, device_coap.Version, d)
}
