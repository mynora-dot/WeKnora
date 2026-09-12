# 登录与注册页面备案信息

在根目录 `.env` 中填写以下公开字段。只影响登录、普通注册和邀请注册页面，登录后的工作区、嵌入聊天和文档站不展示。

```dotenv
ICP_BEIAN_NUMBER=
PUBLIC_SECURITY_BEIAN_NUMBER=
PUBLIC_SECURITY_BEIAN_URL=
PUBLIC_SECURITY_BEIAN_ICON_URL=
```

ICP 字段填写备案平台核发的完整网站备案号，展示后链接至 https://beian.miit.gov.cn/ 。留空不显示。

公安备案使用平台核发的完整编号、HTML 代码中的 HTTPS 查询链接以及官方图标，三项齐全且地址有效才展示。图标可以使用 HTTPS 地址，也可以使用本站绝对路径（例如 `/beian.png`）；使用本站路径时需将实际官方图标放入 `frontend/public/` 后构建，或挂载到容器 `/usr/share/nginx/html/beian.png`。不要填写示例号码或使用自制图标。链接不接受脚本、data 协议或带账号密码的 URL。未配置任何有效项目时页脚不占空间。

首次部署必须构建或使用包含本功能的前端镜像。后续修改 `.env` 后运行：

```sh
docker compose up -d --no-deps --force-recreate frontend
```

浏览器刷新后即可显示新值，无需重新构建镜像。仅执行 `docker compose restart` 不会重新注入 `.env` 的新值。`config.js` 使用现有 Nginx 的禁止缓存复用策略；外部 CDN 不应长期缓存此文件。

本地开发执行 `cd frontend && npm run dev`，Vite 仅将根目录 `.env` 中上述四个字段注入运行时配置；修改后重启开发服务。静态构建或其他托管方式需自行在运行时 `config.js` 中提供同名字段。本功能不读取或公开 `.env` 的其他私密字段。

依据：[通信管理局备案说明](https://gsca.miit.gov.cn/bsfw/bszn/art/2020/art_82d75f07581447f5a8cf74db402554f2.html)、[公安机关备案说明](https://ga.sz.gov.cn/ZT/HLWSYSQ/)。
