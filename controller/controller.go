package controller

import (
	"APCS/service"
)

func Storage() {
}
func Release() {
}
func Sort() {
}
func OverallSort() {
}

func Controller() error {

	err := service.Service.InitService()

	if err != nil {
		return err
	}
	return nil

}
