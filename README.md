# RepoInsight

GitHub 项目分析基于 FastAPI + Streamlit + PostgreSQL，自动爬取、分析并展示 GitHub 项目

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

- **可视化前端**
  - Streamlit 前端，支持项目搜索、热门项目、已分析项目等多页面切换
  - 实时展示AI分析结果

---

## 📁 代码结构

```
repoinsight/
├── app/
│   ├── main.py                # FastAPI 主入口
│   ├── config.py              # 配置管理
│   ├── database.py            # 数据库连接
│   │   ├── repository.py
│   │   ├── ai_analysis.py
│   │   └── crawl_history.py
│   ├── api/
│   │   ├── __init__.py
│   │   └── routes/
│   │       ├── __init__.py
│   │       ├── repositories.py
│   │       └── analysis.py
│   └── web/
│       └── app.py             # Streamlit 前端
├── requirements.txt           # Python依赖
├── Dockerfile                 # Docker镜像构建
├── docker-compose.yml         # 数据库部署
├── schema.sql                 # 数据库表结构
├── config.yml                 # 示例配置
└── README.md
```

---

## ⚙️ 安装与部署

### 方式一：Docker 一键部署（推荐）

1. **克隆项目**
   ```bash
   git clone https://github.com/yourusername/repoinsight.git
   cd repoinsight
   ```

2. **配置环境变量**
   - 新建 `.env` 文件，内容如下（用你自己的Token替换）：
     ```
     GITHUB_TOKEN=你的github_token
     DEEPSEEK_API_KEY=你的deepseek_api_key
     DB_HOST=localhost
     DB_PORT=5432
     DB_USER=postgres
     DB_PASSWORD=postgres
     DB_NAME=repoinsight
     APP_NAME=RepoInsight
     DEBUG=False
     API_PREFIX=/api/v1
     ```

3. **启动服务**
   ```bash
   docker-compose up -d
   ```

4. **访问服务**
   - FastAPI API文档: [http://localhost:8000/docs](http://localhost:8000/docs)
   - Streamlit前端: [http://localhost:8501](http://localhost:8501)

---

### 方式二：本地开发环境

1. 安装依赖
   ```bash
   pip install -r requirements.txt
   ```

2. 初始化数据库（确保PostgreSQL已启动，并执行 `schema.sql`）
   ```bash
   psql -U postgres -d repoinsight -f schema.sql
   ```

3. 启动后端API
   ```bash
   uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
   ```

4. 启动前端
   ```bash
   streamlit run app/web/app.py
   ```

---

## 🖥️ 使用说明

- **项目搜索**：输入关键词，快速查找相关GitHub项目
- **热门项目**：按star数或更新时间展示热门项目
- **已分析项目**：只展示有AI分析结果的项目
- **AI分析**：点击"分析项目"按钮，自动生成并展示AI分析内容

---

## 📝 其他说明

- 所有敏感配置建议通过 `.env` 文件管理
- 数据库结构详见 `schema.sql`
- 支持自定义扩展API和前端页面

---

如有问题或建议，欢迎提issue或联系作者！

