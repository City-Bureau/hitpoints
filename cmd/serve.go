package cmd

import (
	"fmt"
	"log"
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
		port, cronSpec, ssl, sslCert, sslKey := parseBaseArgs(cmd)
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
		port, cronSpec, ssl, sslCert, sslKey := parseBaseArgs(cmd)
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
		port, cronSpec, ssl, sslCert, sslKey := parseBaseArgs(cmd)
		accessKeyID, _ := cmd.Flags().GetString("aws-access-key-id")
		secretAccessKey, _ := cmd.Flags().GetString("aws-secret-access-key")
		s3Bucket, _ := cmd.Flags().GetString("s3-bucket")
		useEnv, _ := cmd.Flags().GetBool("aws-env")
		hitStorage, err := storage.NewS3Storage(accessKeyID, secretAccessKey, s3Bucket, useEnv)

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
	rootCmd.PersistentFlags().String("ssl-crt", "", "SSL certificate file path")
	rootCmd.PersistentFlags().String("ssl-key", "", "SSL key file path")

	fileCmd.Flags().StringP("filepath", "f", "/tmp", "Directory path for writing output files")

	azureCmd.Flags().StringP("account-name", "n", "", "Azure account name")
	azureCmd.Flags().StringP("account-key", "k", "", "Azure account key")
	azureCmd.Flags().StringP("container", "c", "", "Azure container")

	s3Cmd.Flags().StringP("access-key-id", "i", "", "AWS Access key ID")
	s3Cmd.Flags().StringP("secret-access-key", "k", "", "AWS Secret access key")
	s3Cmd.Flags().StringP("bucket", "b", "", "S3 bucket")
	s3Cmd.Flags().Bool("aws-env", false, "Use env vars for setting credentials")
}

func parseBaseArgs(cmd *cobra.Command) (int, string, bool, string, string) {
	port, err := cmd.Parent().PersistentFlags().GetInt("port")
	if err != nil {
		log.Println(err)
	}
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

func serve(port int, hitStorage storage.HitStorage, cronSpec string, ssl bool, sslCert string, sslKey string) {
	hitServer := server.NewHitServer()

	c := cron.New()
	c.AddFunc(cronSpec, func() {
		err := archiveAndClearCache(&hitServer, hitStorage)
		if err != nil {
			log.Println(err)
		}
	})

	c.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/", hitServer.HandlePixelRequest)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}

	if ssl && sslCert != "" && sslKey != "" {
		// Sping up main HTTP port in goroutine
		go func() {
			err := srv.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}()

		sslSrv := &http.Server{
			Addr:         ":443",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      mux,
			TLSConfig:    nil,
		}

		err := sslSrv.ListenAndServeTLS(sslCert, sslKey)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}
}
