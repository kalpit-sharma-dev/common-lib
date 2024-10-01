package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// Config holds the PostgreSQL connection configuration
type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	DBName      string
	SSLMode     string // SSL mode: "verify-ca" or "verify-full"
	SSLRootCert string
	SSLCert     string
	SSLKey      string
}

// AccessSecret fetches a secret from Google Secret Manager by secret name
func AccessSecret(ctx context.Context, client *secretmanager.Client, secretName string) (string, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	// Access the secret version.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	// Return the secret payload as a string
	return string(result.Payload.Data), nil
}

// FetchCertsFromSecretManager retrieves TLS certs from Google Secret Manager
func FetchCertsFromSecretManager(ctx context.Context, projectID string, caCertSecret, clientCertSecret, clientKeySecret string) (string, string, string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create secret manager client: %v", err)
	}
	defer client.Close()

	caCert, err := AccessSecret(ctx, client, fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, caCertSecret))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch CA cert: %v", err)
	}

	clientCert, err := AccessSecret(ctx, client, fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, clientCertSecret))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch client cert: %v", err)
	}

	clientKey, err := AccessSecret(ctx, client, fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, clientKeySecret))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch client key: %v", err)
	}

	return caCert, clientCert, clientKey, nil
}

// ConnectToDB establishes a connection to the PostgreSQL database using TLS and GORM
func ConnectToDB(cfg Config, caCert, clientCert, clientKey string) (*gorm.DB, error) {
	// Load client cert
	cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
	if err != nil {
		return nil, fmt.Errorf("unable to load client cert and key: %v", err)
	}

	// Load CA cert
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM([]byte(caCert)); !ok {
		return nil, fmt.Errorf("failed to append CA cert to the pool")
	}

	// Setup TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	// Create a custom dialer with the TLS configuration
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	// Register the custom dialer with the Postgres driver
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := tls.DialWithDialer(dialer, network, addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	// Build the DSN for GORM
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Open a GORM database connection with PostgreSQL
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
		Conn:                 &sql.Conn{Driver: &pq.Driver{TLSConfig: tlsConfig}},
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "public.", // schema name
			SingularTable: true,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get raw SQL DB: %v", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// QueryDBWithGORM demonstrates running a GORM query with TLS-secured connection
func QueryDBWithGORM(db *gorm.DB) {
	var results []struct {
		Column1 string
		Column2 int
	}

	// Using GORM to execute a query
	if err := db.Table("your_table").Where("column1 = ?", "some_value").Find(&results).Error; err != nil {
		log.Fatalf("Error executing query: %v", err)
	}

	// Print query results
	for _, result := range results {
		fmt.Printf("Row: column1=%s, column2=%d\n", result.Column1, result.Column2)
	}
}

func main() {
	// Define the database connection configuration
	cfg := Config{
		Host:     "your_host",     // Change to your PostgreSQL host
		Port:     5432,            // Change to your PostgreSQL port
		User:     "your_username", // Change to your PostgreSQL username
		Password: "your_password", // Change to your PostgreSQL password
		DBName:   "your_dbname",   // Change to your PostgreSQL database name
		SSLMode:  "verify-full",   // SSL mode: "verify-ca" or "verify-full"
	}

	ctx := context.Background()

	// Fetch certificates from Google Secret Manager
	projectID := "your-gcp-project-id"
	caCertSecret := "your-ca-cert-secret-id"
	clientCertSecret := "your-client-cert-secret-id"
	clientKeySecret := "your-client-key-secret-id"

	caCert, clientCert, clientKey, err := FetchCertsFromSecretManager(ctx, projectID, caCertSecret, clientCertSecret, clientKeySecret)
	if err != nil {
		log.Fatalf("Error fetching certificates from Secret Manager: %v", err)
	}

	// Connect to the database using GORM
	db, err := ConnectToDB(cfg, caCert, clientCert, clientKey)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	fmt.Println("Successfully connected to the PostgreSQL database with TLS")

	// Query database with GORM
	QueryDBWithGORM(db)
}
