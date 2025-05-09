from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.models.repository import Repository
from app.models.ai_analysis import AIAnalysis

router = APIRouter()

def repo_with_analysis(repo, db):
    analysis = db.query(AIAnalysis).filter(AIAnalysis.url == repo.url).first()
    repo_dict = repo.__dict__.copy()
    if analysis:
        repo_dict['analysis'] = {
            'content': analysis.content,
            'status': analysis.status
        }
    else:
        repo_dict['analysis'] = None
    return repo_dict

@router.get("/repositories")
def get_repositories(
    db: Session = Depends(get_db),
    q: str = Query(None, description="搜索关键词"),
    skip: int = 0,
    limit: int = 20
):
    query = db.query(Repository)
    if q:
        query = query.filter(Repository.full_name.ilike(f"%{q}%"))
    repos = query.offset(skip).limit(limit).all()
    return [repo_with_analysis(r, db) for r in repos]

@router.get("/repositories/top")
def get_top_repositories(
    db: Session = Depends(get_db),
    sort: str = Query("stars", description="排序方式: stars/updated"),
    limit: int = 10
):
    if sort == "stars":
        repos = db.query(Repository).order_by(Repository.stars.desc()).limit(limit).all()
    elif sort == "updated":
        repos = db.query(Repository).order_by(Repository.updated_at.desc()).limit(limit).all()
    else:
        repos = db.query(Repository).limit(limit).all()
    return [repo_with_analysis(r, db) for r in repos]

@router.get("/repositories/{repo_id}")
def get_repository_detail(repo_id: int, db: Session = Depends(get_db)):
    repo = db.query(Repository).filter(Repository.id == repo_id).first()
    if not repo:
        return {"error": "Not found"}
    return repo_with_analysis(repo, db)

@router.get("/repositories/test")
async def test_repo():
    return {"msg": "repositories ok"} 