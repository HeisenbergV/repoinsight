from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    # 数据库配置
    DB_HOST: str = "localhost"
    DB_PORT: int = 5432
    DB_USER: str = "postgres"
    DB_PASSWORD: str = "postgres"
    DB_NAME: str = "repoinsight"

    # GitHub配置
    GITHUB_TOKEN: str

    # DeepSeek AI配置
    DEEPSEEK_API_KEY: str

    # 应用配置
    APP_NAME: str = "RepoInsight"
    DEBUG: bool = False
    API_PREFIX: str = "/api/v1"

    class Config:
        env_file = ".env"

settings = Settings() 