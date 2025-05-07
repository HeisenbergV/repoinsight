-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建仓库表
CREATE TABLE IF NOT EXISTS repository (
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
    is_template BOOLEAN DEFAULT FALSE,
    last_analyzed_at TIMESTAMP WITH TIME ZONE,
    analysis_status VARCHAR(20) DEFAULT 'pending',
    search_keyword VARCHAR(255),
    search_rank INTEGER,
    last_crawled_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_repository_full_name ON repository(full_name);
CREATE INDEX IF NOT EXISTS idx_repository_stars ON repository(stars DESC);
CREATE INDEX IF NOT EXISTS idx_repository_language ON repository(language);
CREATE INDEX IF NOT EXISTS idx_repository_last_pushed_at ON repository(last_pushed_at DESC);
CREATE INDEX IF NOT EXISTS idx_repository_deleted_at ON repository(deleted_at);
CREATE INDEX IF NOT EXISTS idx_repository_analysis_status ON repository(analysis_status);
CREATE INDEX IF NOT EXISTS idx_repository_search_keyword ON repository(search_keyword);
CREATE INDEX IF NOT EXISTS idx_repository_last_analyzed_at ON repository(last_analyzed_at);

-- 创建 AI 分析表
CREATE TABLE IF NOT EXISTS ai_analysis (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    url VARCHAR(255) NOT NULL,
    content TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    error_message TEXT,
    analysis_type VARCHAR(50) DEFAULT 'summary',
    model_version VARCHAR(50),
    tokens_used INTEGER,
    UNIQUE(url)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_ai_analysis_url ON ai_analysis(url);
CREATE INDEX IF NOT EXISTS idx_ai_analysis_status ON ai_analysis(status);
CREATE INDEX IF NOT EXISTS idx_ai_analysis_analysis_type ON ai_analysis(analysis_type);

-- 创建爬取历史表
CREATE TABLE IF NOT EXISTS crawl_history (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    keyword VARCHAR(255) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    total_repos INTEGER DEFAULT 0,
    processed_repos INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'running',
    error_message TEXT
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_crawl_history_keyword ON crawl_history(keyword);
CREATE INDEX IF NOT EXISTS idx_crawl_history_status ON crawl_history(status);
CREATE INDEX IF NOT EXISTS idx_crawl_history_started_at ON crawl_history(started_at);

-- 创建每日推送进度表
CREATE TABLE IF NOT EXISTS daily_push_progress (
    id SERIAL PRIMARY KEY,
    topic VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    last_repo_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(topic, date)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_daily_push_progress_topic_date ON daily_push_progress(topic, date);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 删除已存在的触发器
DROP TRIGGER IF EXISTS update_repository_updated_at ON repository;
DROP TRIGGER IF EXISTS update_ai_analysis_updated_at ON ai_analysis;
DROP TRIGGER IF EXISTS update_crawl_history_updated_at ON crawl_history;

-- 创建新的触发器
CREATE TRIGGER update_repository_updated_at
    BEFORE UPDATE ON repository
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ai_analysis_updated_at
    BEFORE UPDATE ON ai_analysis
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_crawl_history_updated_at
    BEFORE UPDATE ON crawl_history
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 