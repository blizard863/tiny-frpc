//go:build gssh
// +build gssh

package main

import (
	"sync"

	"github.com/gofrp/tiny-frpc/pkg/config"
	v1 "github.com/gofrp/tiny-frpc/pkg/config/v1"
	"github.com/gofrp/tiny-frpc/pkg/gssh"
	"github.com/gofrp/tiny-frpc/pkg/util/log"
)

// GoSSHRun 是 Runner 接口的一种具体实现
type GoSSHRun struct{}

func (r GoSSHRun) Run(commonCfg *v1.ClientCommonConfig, pxyCfg []v1.ProxyConfigurer, vCfg []v1.VisitorConfigurer) {
	goSSHParams := config.ParseFRPCConfigToGoSSHParam(commonCfg, pxyCfg, vCfg)

	log.Infof("proxy total len: %v", len(goSSHParams))

	wg := new(sync.WaitGroup)

	for _, cmd := range goSSHParams {
		wg.Add(1)

		go func(cmd config.GoSSHParam) {
			defer wg.Done()

			log.Infof("start to run: %v", cmd)

			tc, err := gssh.NewTunnelClient(cmd.LocalAddr, cmd.ServerAddr, cmd.SSHExtraCmd)
			if err != nil {
				log.Errorf("new ssh tunnel client error: %v", err)
				return
			}

			err = tc.Start()
			if err != nil {
				log.Errorf("cmd: %v run error: %v", cmd, err)
				return
			}
		}(cmd)
	}

	wg.Wait()

	log.Infof("stopping process calling native ssh to frps, exit...")

}

// 构建标签 runner1 下的 init 函数将 runner1 实现赋值给 runner 全局变量
func init() {
	runner = GoSSHRun{}
}
