package utils_test

import (
	"log"
	"testing"

	"github.com/avian-digital-forensics/auto-processing/pkg/utils"
	"github.com/matryer/is"
)

func TestGetIPADdress(t *testing.T) {
	is := is.New(t)
	ip, err := utils.GetIPAddress()
	is.NoErr(err)
	log.Println(ip)
}
