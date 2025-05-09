from sqlalchemy import Column, Integer, String, Text, DateTime
from sqlalchemy.sql import func
from ..database import Base

class AIAnalysis(Base):
    __tablename__ = "ai_analysis"

    id = Column(Integer, primary_key=True, index=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())
    
    url = Column(String(255), unique=True, nullable=False)
    content = Column(Text)
    status = Column(String(20), default='pending')
    error_message = Column(Text)
    analysis_type = Column(String(50), default='summary')
    model_version = Column(String(50))
    tokens_used = Column(Integer) 