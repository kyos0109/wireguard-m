package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	CookieName     = "authenticated"
	passwdFilePath = ".wgmpasswd"
	fixedUsername  = "wgmadmin"
)

// generateRandomPassword generates a random password.
// It generates `byteLength` random bytes and returns a hex string of length byteLength*2.
func generateRandomPassword(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// EnsureWgmPasswdFile checks if the .wgmpasswd file exists.
// If not, it automatically generates a new 16-character password,
// writes it to the file, logs the credentials, and returns the password.
func EnsureWgmPasswdFile() (string, error) {
	if _, err := os.Stat(passwdFilePath); os.IsNotExist(err) {
		// Generate 8 random bytes -> 16 hex characters.
		defaultPassword, err := generateRandomPassword(8)
		if err != nil {
			return "", err
		}
		// Write the generated password to .wgmpasswd.
		if err := os.WriteFile(passwdFilePath, []byte(defaultPassword), 0644); err != nil {
			return "", err
		}
		log.Printf("Created %s with credentials: username: %s, password: %s", passwdFilePath, fixedUsername, defaultPassword)
		return defaultPassword, nil
	}
	// File exists, read and return the stored password.
	data, err := os.ReadFile(passwdFilePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ShowLogin renders the login page.
func ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// DoLogin processes the login form submission and uses the password from .wgmpasswd for verification.
func DoLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Only accept the fixed username.
	if username != fixedUsername {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Authentication failed"})
		return
	}

	// Obtain or create the .wgmpasswd file and its password.
	storedPassword, err := EnsureWgmPasswdFile()
	if err != nil {
		log.Printf("Error ensuring password file: %v", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Internal error"})
		return
	}

	if password != storedPassword {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Authentication failed"})
		return
	}

	// Successful authentication: set cookie and redirect to dashboard.
	c.SetCookie(CookieName, "true", 3600, "/", "", false, true)
	c.Redirect(http.StatusFound, "/dashboard")
}

// AuthRequired is a middleware that checks if the user is authenticated.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(CookieName)
		if err != nil || cookie != "true" {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}
