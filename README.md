
Mock WAF demo
A demo for WAF server to integrate with BFE

本项目配套 bwi一起使用。
用于演示bwi的使用以及bfe中如何新增新的WAF支持。
包含三部分
- waf server demo
- bfe waf client sdk
- 集成到BFE

# waf server demo
一个WAF server的模拟器。实现如下两个功能
- http 请求检测：目前随机5%的请求会设置为攻击。
- 健康检测：：目前随机1%的请求会设置为检测失败。

执行
- Server side:
go run waf_server_demo.go
- Client side detect:
curl  -v   "http://127.0.0.1:8899/detect"
- Client side health check:
 curl  -v   "http://127.0.0.1:8899/hccheck"

# bfe waf client sdk
实现BWI的waf client sdk。可以集成到BFE中。

# 集成到BFE
只需3步共5行代码，如下：
- 在 bfe/go.mod中新增如下require语句
require	github.com/bfenetworks/bfe-mock-waf v0.1.0

- 在 bfe/bfe_modules/mod_unified_waf/waf_impl/waf_imp_entry.go 中新增如下的import
import 	mockWafSDK  "github.com/bfenetworks/bfe-mock-waf/waf-bfe-sdk"

- 在全局的wafImplDict变量中，增加类似如下的成员:
	"BFEMockWaf": &WafImplMethodBundle{
		NewWafServerWithPoolSize: mockWafSDK.NewWafServerWithPoolSize,
		HealthCheck:              mockWafSDK.HealthCheck,
	},

## 如何处理请求
[参考BFE WAF 模块说明](https://github.com/bfenetworks/bfe/tree/develop/docs/zh_cn/modules/mod_unified_waf)


