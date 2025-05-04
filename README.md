# RepoInsight

一个智能化的 GitHub 项目分析工具，通过 AI 技术自动分析 GitHub 项目，帮助开发者快速了解项目特点、技术栈和应用场景。

## 🌟 核心功能

- **智能项目发现**
  - 定时爬取 GitHub 上指定主题的项目（如 AI、区块链等）
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

## 🚀 使用场景

1. **技术调研**
   - 快速了解某个技术领域的热门项目
   - 分析项目的技术栈和实现方式
   - 发现潜在的技术解决方案

2. **项目选型**
   - 比较同类项目的优缺点
   - 了解项目的活跃度和维护状态
   - 评估项目的技术成熟度

3. **学习研究**
   - 发现优质的开源项目
   - 了解最新的技术趋势
   - 学习项目的最佳实践

4. **市场分析**
   - 追踪特定领域的技术发展
   - 分析项目的应用场景
   - 发现潜在的市场机会

## 🛠 技术栈

- **后端**: Go 1.21
- **数据库**: PostgreSQL 15
- **AI 服务**: DeepSeek API
- **部署**: Docker & Docker Compose
- **API 集成**: GitHub API

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
docker-compose up -d db
docker exec -i repoinsight-db psql -U repoinsight_user -d repoinsight_db < schema.sql
```

4. 启动应用服务：
```bash
docker-compose up -d
```

5. 查看日志：
```bash
docker-compose logs -f
```

### 方式二：本地运行

1. 安装依赖：
```bash
go mod download
```

2. 配置应用：
```bash
cp config.yml.example config.yml
```

3. 初始化数据库：
```bash
psql -U repoinsight_user -d repoinsight_db < schema.sql
```

4. 运行项目：
```bash
go run main.go
```

## 🔧 配置说明

配置文件 `config.yml` 包含以下部分：

```yaml
database:
  host: db
  port: 5432
  user: repoinsight_user
  password: your_password
  name: repoinsight_db

api:
  github:
    token: your_github_token
  deepseek:
    api_key: your_deepseek_api_key

app:
  search_keyword: "topic:ai"
  interval_hours: 24
  max_repos_per_search: 100
```

## ⚠️ 注意事项

1. **API 限制**
   - GitHub API 有请求频率限制
   - DeepSeek API 有调用次数限制
   - 建议合理设置爬取间隔

2. **数据安全**
   - 妥善保管 API Token
   - 定期备份数据库
   - 使用配置文件存储敏感信息

3. **性能优化**
   - 根据服务器配置调整并发数
   - 合理设置数据库连接池
   - 定期清理过期数据

## 📝 许可证

MIT License
