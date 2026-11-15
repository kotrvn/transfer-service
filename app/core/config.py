from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")
    
    database_url: str = "postgresql+asyncpg://postgres:secret@localhost:5432/transfer_db"
    
    # --- Настройки для JWT ---
    secret_key: str = "super_secret_key_change_me_in_production"
    algorithm: str = "HS256"
    access_token_expire_minutes: int = 60 * 24 # Токен живет 1 день (1440 минут)

settings = Settings()