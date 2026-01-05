package main

import (
	"strings"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

// Domain to market mapping
var domainMap = map[string]string{
	"www.example.com": "test",
}

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	types.DefaultVMContext
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{contextID: contextID}
}

type pluginContext struct {
	types.DefaultPluginContext
	contextID uint32
}

// Override types.DefaultPluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	proxywasm.LogInfof("market-header plugin started, contextID: %d", ctx.contextID)
	return types.OnPluginStartStatusOK
}

// Override types.DefaultPluginContext.
func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpMarketHeader{contextID: contextID}
}

type httpMarketHeader struct {
	types.DefaultHttpContext
	contextID uint32
	market    string
}

func (ctx *httpMarketHeader) OnHttpRequestHeaders(int, bool) types.Action {
	authority, err := proxywasm.GetHttpRequestHeader(":authority")
	if err != nil {
		proxywasm.LogWarnf("failed to get :authority header: %v", err)
		return types.ActionContinue
	}

	// Remove port if present
	host := authority
	if idx := strings.Index(authority, ":"); idx != -1 {
		host = authority[:idx]
	}

	// Try exact domain match
	if market, ok := domainMap[host]; ok {
		ctx.market = market
		proxywasm.AddHttpRequestHeader("x-request-market", market)
		proxywasm.AddHttpRequestHeader("x-market", market)
		proxywasm.LogInfof("set market: %s for host: %s", market, host)
		return types.ActionContinue
	}

	// Try subdomain pattern (e.g., "at.nonprod.example.com" -> "at")
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		prefix := parts[0]
		if len(prefix) == 2 && isLowerAlpha(prefix) {
			ctx.market = prefix
			proxywasm.AddHttpRequestHeader("x-request-market", prefix)
			proxywasm.AddHttpRequestHeader("x-market", prefix)
			proxywasm.LogInfof("set market: %s for host: %s", prefix, host)
			return types.ActionContinue
		}
	}

	proxywasm.LogInfof("no market mapping for host: %s", host)
	return types.ActionContinue
}

func (ctx *httpMarketHeader) OnHttpResponseHeaders(int, bool) types.Action {
	proxywasm.AddHttpResponseHeader("x-wasm-filter", "market-header")
	if ctx.market != "" {
		proxywasm.AddHttpResponseHeader("x-market-debug", ctx.market)
	}
	return types.ActionContinue
}

func isLowerAlpha(s string) bool {
	for _, c := range s {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}
