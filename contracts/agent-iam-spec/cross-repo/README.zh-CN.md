[English](README.md) · [简体中文](README.zh-CN.md)

# 跨仓一致性 Fixtures

面向 **AgentIAM**（Agent IAM 系列）第 3–7 部分的语言中立、可执行一致性 fixtures。其目的是让独立实现能够在不阅读其他实现源码的前提下，针对自身已发布的契约面证明互操作性。

本项目落实 AEGIVELA 参考实现 ADR-0016 记录的要求：跨仓 fixture 集合由 **`agent-iam-spec` 拥有**，由**各实现执行**。每个期望结果都源自规范条款，而非厂商契约。

## Fixture 是什么

每个 fixture 是符合 [`fixture.schema.json`](fixture.schema.json) 的 JSON 文档：

| 字段 | 含义 |
|---|---|
| `id` | 稳定的 fixture 标识。 |
| `part` / `clause` | 所针对的 AgentIAM 部分与条款。 |
| `threatRef` | 可选的威胁分类（`T1` PoP 绑定、`T2` 工作负载证明、`T4` 生命周期/撤销、`T7` exchange、`T8` 受众绑定、`I1` instance 状态、`F1` 联邦、`BR` 中转签发）。 |
| `kind` | `positive`（必须接受）或 `negative`（必须拒绝）。 |
| `contract` | 目标契约面与版本，例如 `{"surface": "enrollment", "version": "v3.0"}`。 |
| `given` | 测试装置必须建立的前置状态。 |
| `operation.http` | 要发送的方法、路径、header 与 body。 |
| `expected` | `outcome`、HTTP `status`、可选的稳定 `reasonCode`，以及人类可读的 `assertions`。 |

`expected.reasonCode` 为建议性：实现可以使用自身错误分类，只要其因所述原因拒绝且 `assertions` 成立。

## 运行 fixtures

合规的测试装置：

1. 启动被测实现并建立 `given` 状态；
2. 针对实现的契约面执行 `operation.http`；
3. 将响应与 `expected` 比较（状态、outcome 与 assertions）；
4. 按 `id` 报告通过/失败。

Fixtures 以 HTTP 形态描述但传输无关：测试装置可以把 `operation.http` 映射到任何保持请求语义与契约版本的等价调用。

## 覆盖范围

| Part | Fixtures |
|---|---|
| 第 3 部分 — 身份与认证 | enrollment、challenge 单次使用、Token audience、epoch 失效、PoP 重放、凭据 generation、Registry Context 失败关闭、工作负载证明 Profile 版本、JWKS 轮换 overlap |
| 第 4 部分 — 授权与委托 | decision→grant、deny 阻止 grant、委托不可放大、审批绑定、pre-dispatch 撤销、凭据注入 obligation、obligation 强制 |
| 第 5 部分 — 跨域联邦 | 活动 trust、主体隔离、中转字段伪造、trust 禁用、brokered exchange 成功 |
| 第 6 部分 — 审计与安全事件 | 事件持久化、脱敏、事务一致性 |
| 第 7 部分 — 一致性 | claim 完整性、缺部分、组合 profile |

机器可读索引见 [`manifest.json`](manifest.json)。

## 通过运行证明了什么，不证明什么

通过运行证明实现按所声明条款、在所声明契约版本上接受正向 fixtures 并拒绝负向 fixtures。它**不是**认证，不替代部署特定的安全评审，也不证明超出 AgentIAM 范围的企业能力（HSM/KMS 托管、多区域运行、管理控制台）。

## 新增 fixture

在对应 Part 目录下新增 JSON 文档，登记到 `manifest.json`，并运行 `node conformance/validate.mjs`。每个负向 fixture 必须注明其保护的条款；每个安全关键条款宜至少有一个负向 fixture。
