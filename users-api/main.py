# Users API - Gestión de usuarios con FastAPI + SQLite
import hashlib
import os
import sqlite3

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

# Ruta de la base de datos (Docker la pasa con la variable DB_PATH)
DB_PATH = os.getenv("DB_PATH", "data/users.db")

app = FastAPI(title="Users API")


# ---------- Base de datos ----------

def get_db():
    """Abre una conexión a SQLite. row_factory permite leer filas como diccionarios."""
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn


def init_db():
    """Crea la carpeta y la tabla si todavía no existen."""
    os.makedirs(os.path.dirname(DB_PATH) or ".", exist_ok=True)
    conn = get_db()
    conn.execute("""
        CREATE TABLE IF NOT EXISTS users (
            id         INTEGER PRIMARY KEY AUTOINCREMENT,
            name       TEXT NOT NULL,
            email      TEXT NOT NULL UNIQUE,
            password   TEXT NOT NULL,
            created_at TEXT DEFAULT CURRENT_TIMESTAMP
        )
    """)
    conn.commit()
    conn.close()


def hash_password(password: str) -> str:
    """Nunca guardamos la contraseña en texto plano, guardamos su hash SHA-256."""
    return hashlib.sha256(password.encode()).hexdigest()


init_db()


# ---------- Modelos (lo que recibe la API) ----------

class UserCreate(BaseModel):
    name: str
    email: str
    password: str


class UserUpdate(BaseModel):
    name: str | None = None
    email: str | None = None
    password: str | None = None


# Columnas que devolvemos (sin la contraseña)
COLUMNS = "id, name, email, created_at"


# ---------- Endpoints ----------

@app.get("/")
def health():
    return {"service": "users-api", "status": "ok"}


@app.post("/users", status_code=201)
def create_user(user: UserCreate):
    conn = get_db()
    try:
        cursor = conn.execute(
            "INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
            (user.name, user.email, hash_password(user.password)),
        )
        conn.commit()
    except sqlite3.IntegrityError:
        raise HTTPException(status_code=400, detail="El email ya está registrado")
    finally:
        conn.close()
    return get_user(cursor.lastrowid)


@app.get("/users")
def list_users():
    conn = get_db()
    rows = conn.execute(f"SELECT {COLUMNS} FROM users").fetchall()
    conn.close()
    return [dict(row) for row in rows]


@app.get("/users/{user_id}")
def get_user(user_id: int):
    conn = get_db()
    row = conn.execute(f"SELECT {COLUMNS} FROM users WHERE id = ?", (user_id,)).fetchone()
    conn.close()
    if row is None:
        raise HTTPException(status_code=404, detail="Usuario no encontrado")
    return dict(row)


@app.put("/users/{user_id}")
def update_user(user_id: int, user: UserUpdate):
    get_user(user_id)  # lanza 404 si no existe

    # Solo actualizamos los campos que llegaron
    data = user.model_dump(exclude_none=True)
    if "password" in data:
        data["password"] = hash_password(data["password"])
    if not data:
        raise HTTPException(status_code=400, detail="No se enviaron campos para actualizar")

    sets = ", ".join(f"{field} = ?" for field in data)
    conn = get_db()
    try:
        conn.execute(f"UPDATE users SET {sets} WHERE id = ?", (*data.values(), user_id))
        conn.commit()
    except sqlite3.IntegrityError:
        raise HTTPException(status_code=400, detail="El email ya está registrado")
    finally:
        conn.close()
    return get_user(user_id)


@app.delete("/users/{user_id}")
def delete_user(user_id: int):
    get_user(user_id)  # lanza 404 si no existe
    conn = get_db()
    conn.execute("DELETE FROM users WHERE id = ?", (user_id,))
    conn.commit()
    conn.close()
    return {"message": f"Usuario {user_id} eliminado"}
