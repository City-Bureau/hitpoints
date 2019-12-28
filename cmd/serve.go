package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/City-Bureau/hitpoints/pkg/server"
	"github.com/City-Bureau/hitpoints/pkg/storage"
	cron "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

// TODO: Secret key and mode for working like Pixel Ping (flush results when request is made to endpoint)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Store archive files on local paths",
	Long:  `Serve long...`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		port, cronSpec, ssl, sslCert, sslKey := parseBaseArgs(cmd.Parent())
		pathname, _ := cmd.Flags().GetString("filepath")
		hitStorage, err := storage.NewFileStorage(pathname)

		if err != nil {
			log.Fatal(err)
		}

		serve(port, hitStorage, cronSpec, ssl, sslCert, sslKey)
	},
}

var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Store archive files on Azure blob storage",
	Long:  `Serve long...`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		port, cronSpec, ssl, sslCert, sslKey := parseBaseArgs(cmd.Parent())
		accountName, _ := cmd.Flags().GetString("account-name")
		accountKey, _ := cmd.Flags().GetString("account-key")
		container, _ := cmd.Flags().GetString("container")
		hitStorage, err := storage.NewAzureStorage(accountName, accountKey, container)

		if err != nil {
			log.Fatal(err)
		}

		serve(port, hitStorage, cronSpec, ssl, sslCert, sslKey)
	},
}

var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Store archive files on S3",
	Long:  `Serve long...`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		port, cronSpec, ssl, sslCert, sslKey := parseBaseArgs(cmd.Parent())
		accessKeyID, _ := cmd.Flags().GetString("access-key-id")
		secretAccessKey, _ := cmd.Flags().GetString("secret-access-key")
		region, _ := cmd.Flags().GetString("region")
		s3Bucket, _ := cmd.Flags().GetString("bucket")
		useEnv, _ := cmd.Flags().GetBool("use-env")
		hitStorage, err := storage.NewS3Storage(accessKeyID, secretAccessKey, s3Bucket, region, useEnv)

		if err != nil {
			log.Fatal(err)
		}

		serve(port, hitStorage, cronSpec, ssl, sslCert, sslKey)
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)
	rootCmd.AddCommand(azureCmd)
	rootCmd.AddCommand(s3Cmd)

	rootCmd.PersistentFlags().IntP("port", "p", 8080, "Port that the server should run on")
	rootCmd.PersistentFlags().StringP("cron", "c", "30 * * * *", "Cron expression for when archives should be written")
	rootCmd.PersistentFlags().Bool("ssl", false, "Should SSL be enabled for the server")
	rootCmd.PersistentFlags().String("ssl-cert", "", "SSL certificate file path")
	rootCmd.PersistentFlags().String("ssl-key", "", "SSL key file path")

	fileCmd.Flags().StringP("filepath", "f", "/tmp", "Directory path for writing output files")

	azureCmd.Flags().StringP("account-name", "n", "", "Azure account name")
	azureCmd.Flags().StringP("account-key", "k", "", "Azure account key")
	azureCmd.Flags().StringP("container", "c", "", "Azure container")

	s3Cmd.Flags().StringP("access-key-id", "i", "", "AWS Access key ID")
	s3Cmd.Flags().StringP("secret-access-key", "k", "", "AWS Secret access key")
	s3Cmd.Flags().StringP("region", "r", "us-east-1", "AWS region")
	s3Cmd.Flags().StringP("bucket", "b", "", "S3 bucket")
	s3Cmd.Flags().BoolP("use-env", "e", false, "Use env vars for setting credentials")
}

func parseBaseArgs(cmd *cobra.Command) (int, string, bool, string, string) {
	port, _ := cmd.PersistentFlags().GetInt("port")
	cronSpec, _ := cmd.PersistentFlags().GetString("cron")
	ssl, _ := cmd.PersistentFlags().GetBool("ssl")
	sslCert, _ := cmd.PersistentFlags().GetString("ssl-cert")
	sslKey, _ := cmd.PersistentFlags().GetString("ssl-key")

	return port, cronSpec, ssl, sslCert, sslKey
}

func archiveAndClearCache(hitServer *server.HitServer, hitStorage storage.HitStorage) error {
	hitMap := hitServer.CacheItems()
	// Exit if cache is currently empty
	if len(hitMap) == 0 {
		return nil
	}

	err := hitStorage.Archive(hitMap)
	if err != nil {
		return err
	}

	hitServer.ClearCache()
	return nil
}

func serverFromMux(mux *http.ServeMux) *http.Server {
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	u := r.URL
	host, _, _ := net.SplitHostPort(r.Host)
	u.Host = net.JoinHostPort(host, "443")
	u.Scheme = "https"
	http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
}

func serve(port int, hitStorage storage.HitStorage, cronSpec string, ssl bool, sslCert string, sslKey string) {
	hitServer := server.NewHitServer()

	c := cron.New()
	c.AddFunc(cronSpec, func() {
		log.Println("Archiving...")
		err := archiveAndClearCache(&hitServer, hitStorage)
		if err != nil {
			log.Println(err)
		}
	})

	c.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/", hitServer.HandlePixelRequest)
	srv := serverFromMux(mux)

	// TODO: Let's Encrypt configuration https://godoc.org/golang.org/x/crypto/acme/autocert
	if ssl && sslCert != "" && sslKey != "" {
		// Spinning up main HTTP port in goroutine
		go func() {
			redirectMux := http.NewServeMux()
			redirectMux.HandleFunc("/", redirectToHTTPS)
			redirectSrv := serverFromMux(redirectMux)
			redirectSrv.Addr = fmt.Sprintf(":%d", port)
			err := redirectSrv.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}()

		srv.Addr = ":443"
		srv.TLSConfig = &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		}
		err := srv.ListenAndServeTLS(sslCert, sslKey)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		srv.Addr = fmt.Sprintf(":%d", port)
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}
}
