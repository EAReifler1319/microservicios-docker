// Products API - Gestión de productos con Express + SQLite
const express = require("express");
const { DatabaseSync } = require("node:sqlite"); // SQLite ya viene incluido en Node 22+
const fs = require("node:fs");
const path = require("node:path");

// Ruta de la base de datos (Docker la pasa con la variable DB_PATH)
const DB_PATH = process.env.DB_PATH || "data/products.db";
const PORT = 8003;

// ---------- Base de datos ----------

// Crear la carpeta si no existe y abrir la base de datos
fs.mkdirSync(path.dirname(DB_PATH), { recursive: true });
const db = new DatabaseSync(DB_PATH);

// Crear la tabla si todavía no existe
db.exec(`
  CREATE TABLE IF NOT EXISTS products (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    description TEXT,
    price       REAL NOT NULL,
    stock       INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT DEFAULT CURRENT_TIMESTAMP
  )
`);

function findProduct(id) {
  return db.prepare("SELECT * FROM products WHERE id = ?").get(id);
}

// ---------- API ----------

const app = express();
app.use(express.json()); // permite leer el body en formato JSON

app.get("/", (req, res) => {
  res.json({ service: "products-api", status: "ok" });
});

// Registrar producto
app.post("/products", (req, res) => {
  const { name, description = "", price, stock = 0 } = req.body;
  if (!name || price === undefined) {
    return res.status(400).json({ detail: "Los campos name y price son obligatorios" });
  }
  const result = db
    .prepare("INSERT INTO products (name, description, price, stock) VALUES (?, ?, ?, ?)")
    .run(name, description, price, stock);
  res.status(201).json(findProduct(result.lastInsertRowid));
});

// Listar productos
app.get("/products", (req, res) => {
  res.json(db.prepare("SELECT * FROM products").all());
});

// Consultar un producto por ID
app.get("/products/:id", (req, res) => {
  const product = findProduct(req.params.id);
  if (!product) return res.status(404).json({ detail: "Producto no encontrado" });
  res.json(product);
});

// Actualizar producto (solo los campos que lleguen)
app.put("/products/:id", (req, res) => {
  const product = findProduct(req.params.id);
  if (!product) return res.status(404).json({ detail: "Producto no encontrado" });

  // Mezclamos lo que ya existía con lo nuevo
  const updated = { ...product, ...req.body };
  db.prepare("UPDATE products SET name = ?, description = ?, price = ?, stock = ? WHERE id = ?")
    .run(updated.name, updated.description, updated.price, updated.stock, product.id);
  res.json(findProduct(product.id));
});

// Eliminar producto
app.delete("/products/:id", (req, res) => {
  const product = findProduct(req.params.id);
  if (!product) return res.status(404).json({ detail: "Producto no encontrado" });
  db.prepare("DELETE FROM products WHERE id = ?").run(product.id);
  res.json({ message: `Producto ${product.id} eliminado` });
});

app.listen(PORT, () => {
  console.log(`Products API escuchando en el puerto ${PORT}`);
});
