// Auth API - Autenticación (login) con Java + Javalin + SQLite
import io.javalin.Javalin;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.MessageDigest;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.ArrayList;
import java.util.HexFormat;
import java.util.List;
import java.util.Map;

public class Main {

    // Ruta de la base de datos (Docker la pasa con la variable DB_PATH)
    static String DB_PATH = System.getenv().getOrDefault("DB_PATH", "data/auth.db");

    // Lo que recibe /login y /register en el body JSON
    record Credentials(String username, String password) {}

    // ---------- Base de datos ----------

    static Connection getDb() throws SQLException {
        return DriverManager.getConnection("jdbc:sqlite:" + DB_PATH);
    }

    static void initDb() throws Exception {
        Path parent = Path.of(DB_PATH).getParent();
        if (parent != null) Files.createDirectories(parent);

        try (Connection conn = getDb(); Statement st = conn.createStatement()) {
            // Tabla de credenciales válidas
            st.execute("""
                CREATE TABLE IF NOT EXISTS credentials (
                    id            INTEGER PRIMARY KEY AUTOINCREMENT,
                    username      TEXT NOT NULL UNIQUE,
                    password_hash TEXT NOT NULL,
                    created_at    TEXT DEFAULT CURRENT_TIMESTAMP
                )""");
            // Registro de cada intento de login
            st.execute("""
                CREATE TABLE IF NOT EXISTS login_logs (
                    id         INTEGER PRIMARY KEY AUTOINCREMENT,
                    username   TEXT NOT NULL,
                    success    INTEGER NOT NULL,
                    created_at TEXT DEFAULT CURRENT_TIMESTAMP
                )""");
            // Usuario de prueba (OR IGNORE = no falla si ya existe)
            st.execute("INSERT OR IGNORE INTO credentials (username, password_hash) VALUES ('admin', '"
                    + sha256("admin123") + "')");
        }
    }

    // Nunca guardamos la contraseña en texto plano, guardamos su hash SHA-256
    static String sha256(String text) throws Exception {
        byte[] hash = MessageDigest.getInstance("SHA-256").digest(text.getBytes(StandardCharsets.UTF_8));
        return HexFormat.of().formatHex(hash);
    }

    // ---------- API ----------

    public static void main(String[] args) throws Exception {
        initDb();
        Javalin app = Javalin.create();

        // Estado de la API
        app.get("/health", ctx -> ctx.json(Map.of("service", "auth-api", "status", "ok")));

        // Registrar credenciales nuevas
        app.post("/register", ctx -> {
            Credentials c = ctx.bodyAsClass(Credentials.class);
            if (c.username() == null || c.password() == null) {
                ctx.status(400).json(Map.of("detail", "Los campos username y password son obligatorios"));
                return;
            }
            try (Connection conn = getDb();
                 PreparedStatement ps = conn.prepareStatement(
                         "INSERT INTO credentials (username, password_hash) VALUES (?, ?)")) {
                ps.setString(1, c.username());
                ps.setString(2, sha256(c.password()));
                ps.executeUpdate();
                ctx.status(201).json(Map.of("message", "Usuario " + c.username() + " registrado"));
            } catch (SQLException e) {
                ctx.status(400).json(Map.of("detail", "El usuario ya existe"));
            }
        });

        // Login: valida las credenciales y guarda el intento
        app.post("/login", ctx -> {
            Credentials c = ctx.bodyAsClass(Credentials.class);
            if (c.username() == null || c.password() == null) {
                ctx.status(400).json(Map.of("detail", "Los campos username y password son obligatorios"));
                return;
            }
            try (Connection conn = getDb()) {
                // 1. Buscar si existe ese usuario con ese hash
                PreparedStatement ps = conn.prepareStatement(
                        "SELECT id FROM credentials WHERE username = ? AND password_hash = ?");
                ps.setString(1, c.username());
                ps.setString(2, sha256(c.password()));
                boolean valid = ps.executeQuery().next();

                // 2. Guardar el intento en login_logs
                PreparedStatement log = conn.prepareStatement(
                        "INSERT INTO login_logs (username, success) VALUES (?, ?)");
                log.setString(1, c.username());
                log.setInt(2, valid ? 1 : 0);
                log.executeUpdate();

                // 3. Responder
                if (valid) {
                    ctx.json(Map.of("valid", true, "message", "Credenciales válidas"));
                } else {
                    ctx.status(401).json(Map.of("valid", false, "message", "Credenciales inválidas"));
                }
            }
        });

        // Historial de intentos de login
        app.get("/logins", ctx -> {
            List<Map<String, Object>> logs = new ArrayList<>();
            try (Connection conn = getDb();
                 ResultSet rs = conn.createStatement().executeQuery(
                         "SELECT id, username, success, created_at FROM login_logs ORDER BY id DESC")) {
                while (rs.next()) {
                    logs.add(Map.of(
                            "id", rs.getInt("id"),
                            "username", rs.getString("username"),
                            "success", rs.getInt("success") == 1,
                            "created_at", rs.getString("created_at")));
                }
            }
            ctx.json(logs);
        });

        app.start(8001);
    }
}
