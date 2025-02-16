package main

import (
	"fmt"

	"github.com/XigmaDev/huawei"
)

func main() {
	h := huawei.NewHuawei("http://192.168.8.1")

	if err := h.Login("admin", "admin"); err != nil {
		fmt.Println("Login failed:", err)
		return
	}

	fmt.Print(h.SendSMS("done ", "+989123456789"))

}
