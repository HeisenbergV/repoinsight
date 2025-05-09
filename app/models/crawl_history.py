from sqlalchemy import Column, Integer, String, Text, DateTime
from sqlalchemy.sql import func
from ..database import Base

class CrawlHistory(Base):
    __tablename__ = "crawl_history"

    id = Column(Integer, primary_key=True, index=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    keyword = Column(String(255), nullable=False)
    started_at = Column(DateTime(timezone=True), nullable=False)
    completed_at = Column(DateTime(timezone=True))
    total_repos = Column(Integer, default=0)
    processed_repos = Column(Integer, default=0)
    status = Column(String(20), default='running')
    error_message = Column(Text) 