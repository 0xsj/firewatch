from fastapi import FastAPI
from lib.monitoring.logging import configure_logging, get_logger, set_correlation_id

configure_logging("development")
logger = get_logger(__name__)

app = FastAPI(
    title="py-core",
    description="Python API",
    version="0.1.0"
)

@app.get("/")
async def root():
    set_correlation_id()
    logger.info("Root endpoint accessed", user_type="anonymous")
    return {"message": "Hello world", "status": "running"}

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app", 
        host="0.0.0.0", 
        port=8000, 
        reload=True 
    )