package main

import "fmt"

func main() {
	h := NewHuawei("http://192.168.8.1")

	if err := h.Login("admin", "admin"); err != nil {
		fmt.Println("Login failed:", err)
		return
	}

	fmt.Print(h.SendSMS("done ", ""))

}
