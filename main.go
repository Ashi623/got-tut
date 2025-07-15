// package main

// import (
//     "github.com/gofiber/fiber/v2"
// )

// func main() {
//     app := fiber.New()

//     app.Get("/", func(c *fiber.Ctx) error {
//         return c.SendString("Hello, Fiber!")
//     })

//     app.Listen(":3000")
// }


package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

var DB *sql.DB

const (
	HOST     ="dpg-d1j2h06mcj7s73a498b0-a.singapore-postgres.render.com"
	PORT     = "5432"
	USERNAME = "database_nghd_user"
	PASSWORD = "Nl8KrS3s9jDyrblzW9hJYCTNgjtdpnWD"
	DBNAME   = "database_nghd"
)

func main() {
	if err := connectDB(); err != nil {
		log.Fatal("❌ DB Connection Failed:", err)
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("✅ Go app is successfully deployed on Render!")
	})

	app.Post("/users", getAllUsers)
	app.Post("/signin", signInHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	log.Printf("🚀 Server running on port %s", port)
	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func connectDB() error {
	var err error
	psql := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require", HOST, PORT, USERNAME, PASSWORD, DBNAME)
	DB, err = sql.Open("postgres", psql)
	if err != nil {
		return err
	}
	if err = DB.Ping(); err != nil {
		return err
	}
	DB.SetConnMaxIdleTime(10 * time.Minute)
	DB.SetConnMaxLifetime(1 * time.Hour)
	fmt.Println("✅ Connected to PostgreSQL")
	return nil
}

// ---------------------- Handlers ----------------------

func signInHandler(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Invalid request format")
	}

	// Fetch userId and password from DB using email
	var userID int
	var storedPassword string

	err := DB.QueryRow(`
		SELECT u.userid, p.password
		FROM users u
		JOIN passwords p ON u.userid = p.userid
		WHERE u.email = $1
	`, req.Email).Scan(&userID, &storedPassword)

	if err != nil {
		return notFound(c, "User not found with this email")
	}

	// Match password
	if req.Password != storedPassword {
		return badRequest(c, "Incorrect password")
	}

	// Success response
	return c.JSON(fiber.Map{
		"message": "Login successful 🎉",
		//"userId":  userID,
		//"email":   req.Email,
	})
}

func getAllUsers(c *fiber.Ctx) error {
	rows, err := DB.Query("SELECT userid, name, email FROM users")
	if err != nil {
		return serverError(c, err.Error())
	}
	defer rows.Close()

	var users []struct {
		UserID int64  `json:"userId"`
		Name   string `json:"name"`
		Email  string `json:"email"`
	}

	for rows.Next() {
		var u struct {
			UserID int64  `json:"userId"`
			Name   string `json:"name"`
			Email  string `json:"email"`
		}
		if err := rows.Scan(&u.UserID, &u.Name, &u.Email); err != nil {
			return serverError(c, err.Error())
		}
		users = append(users, u)
	}

	return c.JSON(users)
}

// ---------------------- DB Helpers ----------------------

func getUserIDByEmail(email string) (int, error) {
	var userID int
	err := DB.QueryRow("SELECT userid FROM users WHERE email = $1", email).Scan(&userID)
	return userID, err
}

func getEmailByUserID(userID int) (string, error) {
	var email string
	err := DB.QueryRow("SELECT email FROM users WHERE userid = $1", userID).Scan(&email)
	return email, err
}

func getPasswordByUserID(userID int) (string, error) {
	var password string
	err := DB.QueryRow("SELECT password FROM passwords WHERE userid = $1", userID).Scan(&password)
	return password, err
}

// ---------------------- Reusable Responses ----------------------

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(400).JSON(fiber.Map{"error": msg})
}

func notFound(c *fiber.Ctx, msg string) error {
	return c.Status(404).JSON(fiber.Map{"error": msg})
}

func serverError(c *fiber.Ctx, msg string) error {
	return c.Status(500).JSON(fiber.Map{"error": msg})
}
