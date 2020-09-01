/*
Copyright © 2020 AVIAN DIGITAL FORENSICS <sja@avian.dk>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/avian-digital-forensics/auto-processing/cmd/avian/cmd/queue"
	api "github.com/avian-digital-forensics/auto-processing/pkg/avian-api"
	"github.com/avian-digital-forensics/auto-processing/pkg/datastore/tables"
	"github.com/avian-digital-forensics/auto-processing/pkg/services"
	"github.com/avian-digital-forensics/auto-processing/pkg/utils"

	"github.com/gorilla/handlers"
	"github.com/pacedotdev/oto/otohttp"
	ps "github.com/simonjanss/go-powershell"
	"github.com/simonjanss/go-powershell/backend"
	"github.com/spf13/cobra"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "HTTP-service for the queuing component",
	Long: `A http-service for the queue that communicates
with the backend and the running Nuix-scripts.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(); err != nil {
			fmt.Fprintf(os.Stderr, "could not start backend-service: %v\n", err)
		}
	},
}

var (
	address string // Address for http to listen on
	port    string // Port for http to listen on
	debug   bool   // To debug the service
	dbName  string // name for the SQLite-db
)

func init() {
	rootCmd.AddCommand(serviceCmd)

	serviceCmd.Flags().StringVar(&address, "address", "0.0.0.0", "address to listen on")
	serviceCmd.Flags().StringVar(&port, "port", "8080", "port for HTTP to listen on")
	serviceCmd.Flags().BoolVar(&debug, "debug", false, "for debugging")
	serviceCmd.Flags().StringVar(&dbName, "db", "avian.db", "path to sqlite database")
}

func run() error {
	// Set the server-address for HTTP
	if os.Getenv("AVIAN_ADDRESS") == "" {
		ip, err := utils.GetIPAddress()
		if err != nil {
			return fmt.Errorf("Cannot get ip-address: %v", err)
		}
		log.Printf("Warning: no env-variable found for: AVIAN_ADDRESS - using address : %s ", ip)
		address = ip
	} else {
		address = os.Getenv("AVIAN_ADDRESS")
	}

	// Set the server-port for HTTP
	if os.Getenv("AVIAN_PORT") == "" {
		log.Printf("Warning: no env-variable found for: AVIAN_PORT - using port : %s", port)
	} else {
		port = os.Getenv("AVIAN_PORT")
	}

	// Connect to the database
	db, err := gorm.Open("sqlite3", dbName)
	if err != nil {
		return fmt.Errorf("failed to connect database : %v", err)
	}

	// If debug is true, enable logmode
	db.LogMode(debug)

	// Migrate the db-tables
	if err := tables.Migrate(db); err != nil {
		return err
	}

	// Index the db-tables
	if err := tables.Index(db); err != nil {
		return err
	}

	// Create a powershell-shell for remote connections
	shell, err := ps.New(&backend.Local{})
	if err != nil {
		return fmt.Errorf("unable to create powershell-process : %v", err)
	}

	// start the queue
	queue := queue.New(db, shell, fmt.Sprintf("http://%s:%s/oto/", address, port))
	go queue.Start()

	// Create a oto-server
	server := otohttp.NewServer()

	// Register our services
	api.RegisterServerService(server, services.NewServerService(db, shell))
	api.RegisterNmsService(server, services.NewNmsService(db))
	api.RegisterRunnerService(server, services.NewRunnerService(db, shell))

	// Handle our oto-server @ /oto
	http.Handle("/oto/", server)

	// Create our CORS-handlers
	corsOrigins := handlers.AllowedOrigins([]string{"*"})
	corsMethods := handlers.AllowedMethods([]string{"HEAD", "POST", "GET", "DELETE", "PATCH", "PUT", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{
		"Accept",
		"Authorization",
		"Content-Type",
		"User-Agent",
	})

	// Create our HTTP-server
	srv := &http.Server{
		Handler: handlers.CORS(corsOrigins, corsMethods, corsHeaders)(server),
		Addr:    fmt.Sprintf("%s:%s", address, port),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("http-service listening at http://%s:%s\n", address, port)
	return srv.ListenAndServe()
}
