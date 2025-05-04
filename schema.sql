-- 创建数据库（如果不存在）
CREATE DATABASE repoinsight_db;

-- 连接到数据库
\c repoinsight_db;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建仓库表
CREATE TABLE IF NOT EXISTS repositories (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    full_name VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(255) NOT NULL,
    stars INTEGER DEFAULT 0,
    forks INTEGER DEFAULT 0,
    language VARCHAR(50),
    topics TEXT,
    readme TEXT,
    ai_analysis TEXT,
    last_pushed_at TIMESTAMP WITH TIME ZONE,
    is_archived BOOLEAN DEFAULT FALSE,
    license VARCHAR(100),
    default_branch VARCHAR(100),
    open_issues INTEGER DEFAULT 0,
    watchers INTEGER DEFAULT 0,
    size INTEGER DEFAULT 0,
    has_issues BOOLEAN DEFAULT TRUE,
    has_projects BOOLEAN DEFAULT TRUE,
    has_wiki BOOLEAN DEFAULT TRUE,
    has_pages BOOLEAN DEFAULT FALSE,
    has_downloads BOOLEAN DEFAULT TRUE,
    is_template BOOLEAN DEFAULT FALSE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_repositories_full_name ON repositories(full_name);
CREATE INDEX IF NOT EXISTS idx_repositories_stars ON repositories(stars DESC);
CREATE INDEX IF NOT EXISTS idx_repositories_language ON repositories(language);
CREATE INDEX IF NOT EXISTS idx_repositories_last_pushed_at ON repositories(last_pushed_at DESC);
CREATE INDEX IF NOT EXISTS idx_repositories_deleted_at ON repositories(deleted_at);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_repositories_updated_at
    BEFORE UPDATE ON repositories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 