# huawei-mini-library

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/xigmaDev/huawei-mini-library)
[![Go Report Card](https://goreportcard.com/badge/github.com/xigmaDev/huawei-mini-library)](https://goreportcard.com/report/github.com/xigmaDev/huawei-mini-library)
[![Issues](https://img.shields.io/github/issues/xigmaDev/huawei-mini-library.svg)](https://github.com/xigmaDev/huawei-mini-library/issues)

Package `huawei` provides a lightweight interface for communicating with the Huawei E5336B dongle.

This package simplifies interaction with the Huawei E5336B dongle by offering a minimal API for device communication, SMS management, and connection monitoring.

It provides:

 * `NewHuawei` to create a new connection to the device.
 * `Connect` and `Disconnect` to manage the connection lifecycle.
 * `GetDeviceInfo` to retrieve detailed device information.
 * `SendSMS` to send SMS messages through the dongle.


thanks to [max246](https://github.com/max246)
