package main

import (
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

// Simple domain to market mapping - avoid complex string operations
var domainMap = map[string]string{
	"www.example.com": "test",
}

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	types.DefaultVMContext
}

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	types.DefaultPluginContext
}

func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	proxywasm.LogInfo("market-header plugin started")
	return types.OnPluginStartStatusOK
}

func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{}
}

type httpContext struct {
	types.DefaultHttpContext
}

func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	host, err := proxywasm.GetHttpRequestHeader(":authority")
	if err != nil {
		proxywasm.LogWarn("failed to get host")
		return types.ActionContinue
	}

	// Simple lookup - avoid string operations that might cause issues
	if market, ok := domainMap[host]; ok {
		_ = proxywasm.AddHttpRequestHeader("x-market", market)
		_ = proxywasm.AddHttpRequestHeader("x-request-market", market)
		proxywasm.LogInfof("market=%s host=%s", market, host)
	}

	return types.ActionContinue
}

func (ctx *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	_ = proxywasm.AddHttpResponseHeader("x-wasm-filter", "market-header")
	return types.ActionContinue
}
