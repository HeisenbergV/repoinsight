# RepoInsight
一个智能化的 GitHub 项目分析工具，通过 AI 技术自动分析 GitHub 项目，帮助开发者快速了解项目特点、技术栈和应用场景。

## 🌟 核心功能

- **智能项目发现**
  - 定时爬取 GitHub 上指定主题的项目
  - 支持自定义搜索关键词和过滤条件
  - 自动获取项目 README 和描述信息

- **AI 智能分析**
  - 使用 AI 分析项目 README 内容
  - 自动总结项目特点、功能和技术栈
  - 识别项目的潜在应用场景
  - 生成中文分析报告

- **数据存储与管理**
  - 使用 PostgreSQL 存储项目信息和分析结果
  - 支持数据持久化和备份
  - 提供完整的数据查询能力


## 📦 安装部署

### 方式一：Docker 部署（推荐）

1. 克隆项目：
```bash
git clone https://github.com/yourusername/repoinsight.git
cd repoinsight
```

2. 配置应用：
```bash
cp config.yml.example config.yml
```
编辑 `config.yml` 文件，填入你的 GitHub Token 和 DeepSeek API Key。

3. 初始化数据库：
```bash
docker-compose up -d
```

4. 启动应用服务

