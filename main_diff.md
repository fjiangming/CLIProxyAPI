# dev 分支与 main 分支差异分析报告

> 生成时间：2026-04-12  
> 当前分支：`dev`  
> 对比基准：`origin/main` (`c4459c43`)  
> 合并基点（merge-base）：`c4459c43`（即 main HEAD，dev 分支从此点分叉）  
> Fork 上游仓库：`router-for-me/CLIProxyAPI`

---

## 1. 总览

| 指标 | 值 |
|------|----|
| dev 领先 main 的提交数 | 13 (包括重定向逻辑提交) |
| 修改文件数 | 5 |
| 新增文件数 | 2 |
| 总行数变化 | +545 / -61 (约值) |

### 变更文件清单

| 状态 | 文件路径 | 变更行数 | 功能分类 |
|------|----------|----------|----------|
| **M** | `.github/workflows/docker-image.yml` | +159 / -61 | CI/CD |
| **M** | `Dockerfile` | +3 | 构建 |
| **M** | `README_CN.md` | +248 | 文档 |
| **M** | `docker-compose.yml` | +1 / -1 | 部署 |
| **M** | `internal/api/server.go` | +2 | 核心代码 |
| **A** | `internal/api/user_management.go` | +76 | 核心代码（新增） |
| **A** | `user-oauth.html` | +117 | 前端页面（新增） |

---

## 2. 各文件变更详情

### 2.1 `.github/workflows/docker-image.yml` — CI/CD 流水线重构

**变更性质**：大幅改写

**核心改动**：
1. **镜像仓库迁移**：DockerHub (`eceasy/cli-proxy-api`) → GHCR (`ghcr.io/fjiangming/cli-proxy-api`)
2. **新增 dev 分支构建触发**：`on.push.branches: [dev]`，原仅触发 tag push
3. **认证方式变更**：`DOCKERHUB_USERNAME/TOKEN` → `github.actor/GITHUB_TOKEN`
4. **标签策略调整**：tag push 构建 `latest` + 版本号，dev push 构建 `dev` 标签
5. **添加 GHCR 权限声明**：每个 job 增加 `permissions: contents/packages`
6. **清理逻辑简化**：不支持的标签删除 API 废弃。

---

### 2.2 `Dockerfile` — 构建配置

**变更性质**：小幅新增（+3 行）

```diff
+RUN mkdir -p /CLIProxyAPI/static
+COPY user-oauth.html /CLIProxyAPI/static/user-oauth.html
```

**说明**：将 `user-oauth.html` 打包到 Docker 镜像的 static 目录中。

---

### 2.3 `README_CN.md` — 中文部署文档

**变更性质**：大幅新增（+248 行）

**新增内容**：
- VPS 基础环境配置、Docker 安装及 Cloudflare Tunnel 部署指南。CPA 服务起降与密钥保留策略等。

---

### 2.4 `docker-compose.yml` — 部署配置

**变更性质**：微调（1 行）

```diff
-    image: ${CLI_PROXY_IMAGE:-eceasy/cli-proxy-api:latest}
+    image: ${CLI_PROXY_IMAGE:-ghcr.io/fjiangming/cli-proxy-api:dev}
```

**说明**：默认镜像从 DockerHub 切换到 GHCR dev 标签。

---

### 2.5 `internal/api/server.go` — 服务注册入口

**变更性质**：微调（+2 行）

```diff
 	s.registerManagementRoutes()
 }
+
+	s.registerUserManagementRoutes() // [custom] user-level OAuth routes (API Key auth)

 if optionState.keepAliveEnabled {
```

**说明**：在 `NewServer()` 中注册了用于 OAuth 的用户级新路由。

---

### 2.6 `internal/api/user_management.go` — 用户级 OAuth 管理（新增文件）

**功能**：
- 注册 `/user-oauth.html` 及 `/v0/user-oauth/*` 等 API。
- 给各服务下发 OAuth Token 及提供登录流程入口代理（跳开原生管理界面的管控权限拦截）。

---

### 2.7 `user-oauth.html` — 用户 OAuth 前端页面（新增定制页面）

**功能**：
- 基于原生管理页面通过 JavaScript 注入复用页面样式并规避授权检查。
- 隐藏多余导航栏从而达成干净独立的视图环境。
- **登录成功重定向机制**：监听了 `hashchange` 事件，成功跳转后非正式域名（非 `cpa.fjmooo.cn` ）将自动重定向。

---

## 3. 上游合并冲突风险评估及处理指南

> 核心合并原则：**以定制化代码为准（如 GHCR 构建体系、定制路由及 html 覆盖打包等），在解决冲突时优先保留定制改动，并引入和兼容上游的新特性。**

### 3.1 🔴 `.github/workflows/docker-image.yml`
**风险：极高**（必定冲突）
**解决策略**：上游流水线的任何修改都会在合并时产生冲突。请保留现有的 GHCR 定制化及多架构构建策略（**keep ours**）。如果上游更新了重要依赖环境（升级系统、修改内置环境变量等）需要手动从上游改动中抽取过来加入我们。

### 3.2 🟡 `Dockerfile` & `docker-compose.yml`
**风险：中**
**解决策略**：
- `Dockerfile` 若冲突，需手动合并并保证我们新增的 `user-oauth.html` COPY 命令存在即可。
- `docker-compose.yml` 必须保留现有的使用本仓库 GHCR 构建发布镜像的环境变量指代。接纳除 image 环境源修改外的所有其他上游变更。

### 3.3 🟡 `internal/api/server.go`
**风险：中**
**解决策略**：若上游服务入口注册项大改可能导致冲突（我们的代码是在 `registerManagementRoutes()` 附近）。通过确保我们对 `s.registerUserManagementRoutes()` 的单独调用存在且未缺失，即可顺利解决。

### 3.4 🟢 新增内容（HTML 和 GO 代码）
**风险：极低**
**解决策略**：全新编写和隔离解耦的专属文件被重名覆盖的可能性基本没有，除非其依赖项比如底层 `s.mgmt.*` 方法签名或包发生了改变，如有编译报错或调整，按照上游对应的变动修改其对上游内部管理方法的入参出参即可。
