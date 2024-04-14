//go:build nssh
// +build nssh

package main

import (
	"context"
	"sync"

	"github.com/gofrp/tiny-frpc/pkg/config"
	v1 "github.com/gofrp/tiny-frpc/pkg/config/v1"
	"github.com/gofrp/tiny-frpc/pkg/nssh"
	"github.com/gofrp/tiny-frpc/pkg/util/log"
)

// NativeSSHRun 是 Runner 接口的一种具体实现
type NativeSSHRun struct{}

func (r NativeSSHRun) Run(commonCfg *v1.ClientCommonConfig, pxyCfg []v1.ProxyConfigurer, vCfg []v1.VisitorConfigurer) {
	sshCmds := config.ParseFRPCConfigToSSHCmd(commonCfg, pxyCfg, vCfg)

	log.Infof("proxy total len: %v", len(sshCmds))

	closeCh := make(chan struct{})
	wg := new(sync.WaitGroup)

	for _, cmd := range sshCmds {
		wg.Add(1)

		go func(cmd string) {
			defer wg.Done()
			ctx := context.Background()

			log.Infof("start to run: %v", cmd)

			task := nssh.NewCmdWrapper(ctx, cmd, closeCh)
			task.ExecuteCommand(ctx)
		}(cmd)
	}

	wg.Wait()
	close(closeCh)

	log.Infof("stopping process calling native ssh to frps, exit...")
}

// 构建标签 runner1 下的 init 函数将 runner1 实现赋值给 runner 全局变量
func init() {
	runner = NativeSSHRun{}
}
