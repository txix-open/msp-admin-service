package conf_test

import (
	"testing"

	"github.com/txix-open/isp-kit/test/rct"
	"msp-admin-service/conf"
)

func TestDefaultRemoteConfig(t *testing.T) {
	t.Parallel()
	rct.Test(t, "default_remote_config.json", conf.Remote{})
}
