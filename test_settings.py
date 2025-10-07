"""Quick test script for settings"""

from app.core.config import settings


def test_settings():
    """Test that settings load correctly."""
    print("=" * 50)
    print("Testing Settings Configuration")
    print("=" * 50)
    
    # Application settings
    print(f"\n📱 Application:")
    print(f"  Name: {settings.APP_NAME}")
    print(f"  Version: {settings.APP_VERSION}")
    print(f"  Environment: {settings.ENVIRONMENT}")
    print(f"  Debug: {settings.DEBUG}")
    
    # Server settings
    print(f"\n🖥️  Server:")
    print(f"  Host: {settings.HOST}")
    print(f"  Port: {settings.PORT}")
    print(f"  Workers: {settings.WORKERS}")
    
    # Database settings
    print(f"\n🗄️  Database:")
    print(f"  Host: {settings.POSTGRES_HOST}")
    print(f"  Port: {settings.POSTGRES_PORT}")
    print(f"  Database: {settings.POSTGRES_DB}")
    print(f"  User: {settings.POSTGRES_USER}")
    print(f"  URL: {settings.DATABASE_URL}")
    
    # Redis settings
    print(f"\n🔴 Redis:")
    print(f"  Host: {settings.REDIS_HOST}")
    print(f"  Port: {settings.REDIS_PORT}")
    print(f"  DB: {settings.REDIS_DB}")
    print(f"  URL: {settings.REDIS_URL}")
    
    # Security settings
    print(f"\n🔒 Security:")
    print(f"  Algorithm: {settings.ALGORITHM}")
    print(f"  Token Expire: {settings.ACCESS_TOKEN_EXPIRE_MINUTES} min")
    print(f"  Secret Key Length: {len(settings.SECRET_KEY)} chars")
    
    # Environment checks
    print(f"\n🌍 Environment Checks:")
    print(f"  Is Development: {settings.is_development}")
    print(f"  Is Production: {settings.is_production}")
    print(f"  Is Staging: {settings.is_staging}")
    
    print("\n" + "=" * 50)
    print("✅ All settings loaded successfully!")
    print("=" * 50)


if __name__ == "__main__":
    test_settings()