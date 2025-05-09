from sqlalchemy import Column, Integer, String, Text, Boolean, DateTime, ForeignKey
from sqlalchemy.sql import func
from ..database import Base

class Repository(Base):
    __tablename__ = "repository"

    id = Column(Integer, primary_key=True, index=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    deleted_at = Column(DateTime(timezone=True), nullable=True)
    
    full_name = Column(String(255), unique=True, nullable=False)
    name = Column(String(255), nullable=False)
    owner = Column(String(255), nullable=False)
    description = Column(Text)
    url = Column(String(255), nullable=False)
    stars = Column(Integer, default=0)
    forks = Column(Integer, default=0)
    language = Column(String(50))
    topics = Column(Text)
    readme = Column(Text)
    last_pushed_at = Column(DateTime(timezone=True))
    is_archived = Column(Boolean, default=False)
    license = Column(String(100))
    default_branch = Column(String(100))
    open_issues = Column(Integer, default=0)
    watchers = Column(Integer, default=0)
    size = Column(Integer, default=0)
    has_issues = Column(Boolean, default=True)
    has_projects = Column(Boolean, default=True)
    has_wiki = Column(Boolean, default=True)
    has_pages = Column(Boolean, default=False)
    has_downloads = Column(Boolean, default=True)
    is_template = Column(Boolean, default=False)
    last_analyzed_at = Column(DateTime(timezone=True))
    analysis_status = Column(String(20), default='pending')
    search_keyword = Column(String(255))
    search_rank = Column(Integer)
    last_crawled_at = Column(DateTime(timezone=True)) 