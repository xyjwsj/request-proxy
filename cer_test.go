package proxy

import (
	"github.com/xyjwsj/request-proxy/util"
	"testing"
)

func TestRootCer(t *testing.T) {
	certificate := util.NewCertificate()
	certificate.GenerateRootPemFile("platform.hoolai.com")
	certificate.Init()
	certificate.GeneratePem("platform.hoolai.com")
}
