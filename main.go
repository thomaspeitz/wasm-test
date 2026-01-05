package main

import (
	"strings"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

// Domain to market mapping
var domainMap = map[string]string{
	// Test domain
	"www.example.com": "test",
}

type vmContext struct{}

func (*vmContext) OnVMStart(vmConfigurationSize int) types.OnVMStartStatus {
	proxywasm.LogInfo("market-header WASM plugin initialized")
	return types.OnVMStartStatusOK
}

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	types.DefaultPluginContext
}

func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{}
}

type httpContext struct {
	types.DefaultHttpContext
}

func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	// Get the :authority header (host)
	authority, err := proxywasm.GetHttpRequestHeader(":authority")
	if err != nil {
		proxywasm.LogWarnf("failed to get :authority header: %v", err)
		return types.ActionContinue
	}

	// Remove port if present (e.g., "example.com:8080" -> "example.com")
	host := authority
	if idx := strings.Index(authority, ":"); idx != -1 {
		host = authority[:idx]
	}

	// Try exact domain match first
	if market, ok := domainMap[host]; ok {
		ctx.addMarketHeaders(market, host)
		return types.ActionContinue
	}

	// Try nonprod subdomain pattern (e.g., "at.nonprod.example.com" -> "at")
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		prefix := parts[0]
		// Check if it's a 2-letter market code (all lowercase)
		if len(prefix) == 2 && isLowerAlpha(prefix) {
			ctx.addMarketHeaders(prefix, host)
			return types.ActionContinue
		}
	}

	proxywasm.LogInfof("no market mapping found for host: %s", host)
	return types.ActionContinue
}

func (ctx *httpContext) addMarketHeaders(market, host string) {
	if err := proxywasm.AddHttpRequestHeader("x-request-market", market); err != nil {
		proxywasm.LogWarnf("failed to add x-request-market header: %v", err)
	}
	if err := proxywasm.AddHttpRequestHeader("x-market", market); err != nil {
		proxywasm.LogWarnf("failed to add x-market header: %v", err)
	}
	proxywasm.LogInfof("set market header: %s for host: %s", market, host)
}

func isLowerAlpha(s string) bool {
	for _, c := range s {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}
