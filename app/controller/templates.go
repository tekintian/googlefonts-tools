package controller

import "github.com/tekintian/googlefonts-tools/app/templates"

var AppVer = "dev"

func SetAppVer(ver string) {
	AppVer = ver
	templates.AppVer = ver
}
