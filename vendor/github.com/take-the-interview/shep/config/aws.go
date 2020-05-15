package config

import (
	"fmt"
	"os"
)

func GetAWSRegion() (region string) {
	region = os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	if region == "" {
		fmt.Println("You need to set env variable AWS_REGION or AWS_DEFAULT_REGION for this to work.")
		os.Exit(1)
	}
	return
}
