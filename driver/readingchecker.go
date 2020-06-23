// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"math"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/spf13/cast"
)

// checkValueInRange checks value range is valid
func checkValueInRange(valueType dsModels.ValueType, reading interface{}) bool {
	isValid := false

	if valueType == dsModels.String || valueType == dsModels.Bool {
		return true
	}

	if valueType == dsModels.Int8 || valueType == dsModels.Int16 ||
		valueType == dsModels.Int32 || valueType == dsModels.Int64 {
		val := cast.ToInt64(reading)
		isValid = checkIntValueRange(valueType, val)
	}

	if valueType == dsModels.Uint8 || valueType == dsModels.Uint16 ||
		valueType == dsModels.Uint32 || valueType == dsModels.Uint64 {
		val := cast.ToUint64(reading)
		isValid = checkUintValueRange(valueType, val)
	}

	if valueType == dsModels.Float32 || valueType == dsModels.Float64 {
		val := cast.ToFloat64(reading)
		isValid = checkFloatValueRange(valueType, val)
	}

	return isValid
}

func checkUintValueRange(valueType dsModels.ValueType, val uint64) bool {
	var isValid = false
	switch valueType {
	case dsModels.Uint8:
		if val >= 0 && val <= math.MaxUint8 {
			isValid = true
		}
	case dsModels.Uint16:
		if val >= 0 && val <= math.MaxUint16 {
			isValid = true
		}
	case dsModels.Uint32:
		if val >= 0 && val <= math.MaxUint32 {
			isValid = true
		}
	case dsModels.Uint64:
		maxiMum := uint64(math.MaxUint64)
		if val >= 0 && val <= maxiMum {
			isValid = true
		}
	}
	return isValid
}

func checkIntValueRange(valueType dsModels.ValueType, val int64) bool {
	var isValid = false
	switch valueType {
	case dsModels.Int8:
		if val >= math.MinInt8 && val <= math.MaxInt8 {
			isValid = true
		}
	case dsModels.Int16:
		if val >= math.MinInt16 && val <= math.MaxInt16 {
			isValid = true
		}
	case dsModels.Int32:
		if val >= math.MinInt32 && val <= math.MaxInt32 {
			isValid = true
		}
	case dsModels.Int64:
		if val >= math.MinInt64 && val <= math.MaxInt64 {
			isValid = true
		}
	}
	return isValid
}

func checkFloatValueRange(valueType dsModels.ValueType, val float64) bool {
	var isValid = false
	switch valueType {
	case dsModels.Float32:
		if math.Abs(val) >= math.SmallestNonzeroFloat32 && math.Abs(val) <= math.MaxFloat32 {
			isValid = true
		}
	case dsModels.Float64:
		if math.Abs(val) >= math.SmallestNonzeroFloat64 && math.Abs(val) <= math.MaxFloat64 {
			isValid = true
		}
	}
	return isValid
}