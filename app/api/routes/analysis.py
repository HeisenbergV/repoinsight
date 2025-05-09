from fastapi import APIRouter, Depends, Body
from sqlalchemy.orm import Session
from app.database import get_db
from app.models.ai_analysis import AIAnalysis

router = APIRouter()

@router.post("/analysis/analyze")
def analyze_project(
    db: Session = Depends(get_db),
    url: str = Body(..., embed=True)
):
    # 这里只做数据库查询，实际AI分析逻辑可后续补充
    analysis = db.query(AIAnalysis).filter(AIAnalysis.url == url).first()
    if analysis:
        return {"content": analysis.content, "status": analysis.status}
    else:
        return {"content": "暂无分析结果", "status": "pending"}

@router.get("/analysis/test")
async def test_analysis():
    return {"msg": "analysis ok"} 