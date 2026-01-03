下面是我帮你整理后的 Markdown 版 README，结构更清晰，也更符合开源/项目文档的常见规范，直接可用在 GitHub / GitLab。
<img width="1078" height="500" alt="result" src="https://github.com/user-attachments/assets/e8e3edcf-4d57-4c47-9678-9baf414a4ddd" />

⸻

微信小程序 · 家庭菜谱（AI 辅助生成）

一个面向家庭场景的微信小程序，用于 家庭成员协作管理菜谱、订餐与交流，并支持 AI 辅助生成菜谱内容。

⸻

✨ 功能介绍

1️⃣ 家庭（Family）管理
	•	创建家庭
	•	邀请成员加入家庭
	•	移除家庭成员

2️⃣ 菜谱（Recipe）管理
	•	添加菜谱（文字 + 图片）
	•	修改 / 删除菜谱
	•	菜谱列表页
	•	菜谱详情页预览

3️⃣ 订餐功能
	•	为家庭成员预定某一天的：
	•	早餐
	•	午餐
	•	晚餐
	•	菜品来源于家庭菜谱库

4️⃣ 留言板
	•	家庭内部留言
	•	成员之间交流与互动

⸻

🧱 技术栈

前端
	•	JavaScript
	•	微信小程序（WXML / WXSS）

后端
	•	Golang
	•	MySQL

基础设施
	•	Nginx（反向代理）
	•	Docker & Docker Compose

⸻

🚀 部署方式
	•	Docker Compose 一键部署
	•	支持本地开发与线上部署

⸻

📦 使用方式

1️⃣ 前端配置（微信开发者工具）
	1.	将项目根目录导入到 微信开发者工具
	2.	在项目中新增 config.json 文件：

{
  "app_id": "微信小程序的 appId",
  "app_secret": "对应的密钥"
}

	3.	如果不是本地部署，请全局搜索并替换：

localhost → 你的真实域名


⸻

2️⃣ 启动后端服务（Docker）

在已安装 Docker 的情况下：

cd wx_mini_program_golang
docker compose down && docker compose up -d --build

启动后将自动运行：
	•	Golang 后端服务
	•	MySQL
	•	Nginx Proxy Manager

⸻

3️⃣ 配置 Nginx Proxy Manager（非常关键）
	1.	打开管理后台：

http://localhost:81

	2.	默认登录账号：

Email:    admin@example.com
Password: changeme

⚠️ 首次登录会强制修改邮箱和密码

	3.	新增 Proxy Hosts（以本地开发为例）：

配置项	值
Domain Name	localhost
Forward Hostname	my_golang_app（Docker 内部服务名）
Forward Port	8080

	•	若为线上部署：
	•	修改 Domain Name 为真实域名
	•	启用并配置 SSL（推荐）

⸻

4️⃣ 运行小程序
	•	在 微信开发者工具 中点击「编译 / 预览」
	•	登录后即可使用全部功能

⸻

📝 说明
	•	图片上传已支持前端压缩，降低流量消耗
	•	后端可进一步进行图片二次压缩与安全校验
	•	适合家庭私有化部署或小规模使用
