package main

import (
	"flag"
	"fmt"

	"github.com/gofrp/tiny-frpc/pkg/config"
	v1 "github.com/gofrp/tiny-frpc/pkg/config/v1"
	"github.com/gofrp/tiny-frpc/pkg/util"
	"github.com/gofrp/tiny-frpc/pkg/util/log"
	"github.com/gofrp/tiny-frpc/pkg/util/version"
)

func main() {
	var (
		cfgFilePath string
		showVersion bool
	)

	flag.StringVar(&cfgFilePath, "c", "frpc.toml", "path to the configuration file")
	flag.BoolVar(&showVersion, "v", false, "version of tiny-frpc")
	flag.Parse()

	if showVersion {
		fmt.Println(version.Full())
		return
	}

	cfg, proxyCfgs, visitorCfgs, _, err := config.LoadClientConfig(cfgFilePath, true)
	if err != nil {
		log.Errorf("load frpc config error: %v", err)
		return
	}

	_, err = v1.ValidateAllClientConfig(cfg, proxyCfgs, visitorCfgs)
	if err != nil {
		log.Errorf("validate frpc config error: %v", err)
		return
	}

	log.Infof("common cfg: %v, proxy cfg: %v, visitor cfg: %v", util.JSONEncode(cfg), util.JSONEncode(proxyCfgs), util.JSONEncode(visitorCfgs))

	runner.Run(cfg, proxyCfgs, visitorCfgs)
}

type Runner interface {
	Run(commonCfg *v1.ClientCommonConfig, pxyCfg []v1.ProxyConfigurer, vCfg []v1.VisitorConfigurer)
}

type defaultRunner struct{}

func (r defaultRunner) Run(commonCfg *v1.ClientCommonConfig, pxyCfg []v1.ProxyConfigurer, vCfg []v1.VisitorConfigurer) {
	fmt.Println("Running default implementation")
}

var runner Runner = defaultRunner{}
