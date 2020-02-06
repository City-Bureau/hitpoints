package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/City-Bureau/hitpoints/pkg/server"
	"github.com/City-Bureau/hitpoints/pkg/storage"

	cron "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/acme/autocert"
)

// CommandConfig is a wrapper around several base command config parameters
type CommandConfig struct {
	port     int
	cronSpec string
	domain   string
	ssl      bool
}

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Store archive files on local paths",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmdConf := parseBaseArgs(cmd.Parent())
		pathname, _ := cmd.Flags().GetString("filepath")
		hitStorage, err := storage.NewFileStorage(pathname)

		if err != nil {
			log.Fatal(err)
		}

		serve(cmdConf, hitStorage)
	},
}

var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Store archive files on Azure blob storage",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmdConf := parseBaseArgs(cmd.Parent())
		accountName, _ := cmd.Flags().GetString("account-name")
		accountKey, _ := cmd.Flags().GetString("account-key")
		container, _ := cmd.Flags().GetString("container")
		hitStorage, err := storage.NewAzureStorage(accountName, accountKey, container)

		if err != nil {
			log.Fatal(err)
		}

		serve(cmdConf, hitStorage)
	},
}

var s3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Store archive files on S3",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmdConf := parseBaseArgs(cmd.Parent())
		accessKeyID, _ := cmd.Flags().GetString("access-key-id")
		secretAccessKey, _ := cmd.Flags().GetString("secret-access-key")
		region, _ := cmd.Flags().GetString("region")
		s3Bucket, _ := cmd.Flags().GetString("bucket")
		useRole, _ := cmd.Flags().GetBool("role")
		hitStorage, err := storage.NewS3Storage(accessKeyID, secretAccessKey, s3Bucket, region, useRole)

		if err != nil {
			log.Fatal(err)
		}

		serve(cmdConf, hitStorage)
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)
	rootCmd.AddCommand(azureCmd)
	rootCmd.AddCommand(s3Cmd)

	rootCmd.PersistentFlags().IntP("port", "p", 8080, "Port that the server should run on")
	rootCmd.PersistentFlags().StringP("cron", "c", "30 * * * *", "Cron expression for when archives should be written")
	rootCmd.PersistentFlags().StringP("domain", "d", "", "Domain for Let's Encrypt")
	rootCmd.PersistentFlags().Bool("ssl", false, "Should SSL be enabled for the server")

	fileCmd.Flags().StringP("filepath", "f", "/tmp", "Directory path for writing output files")

	azureCmd.Flags().StringP("account-name", "n", "", "Azure account name")
	azureCmd.Flags().StringP("account-key", "k", "", "Azure account key")
	azureCmd.Flags().String("container", "", "Azure container")

	s3Cmd.Flags().StringP("access-key-id", "i", "", "AWS Access key ID")
	s3Cmd.Flags().StringP("secret-access-key", "k", "", "AWS Secret access key")
	s3Cmd.Flags().StringP("region", "r", "us-east-1", "AWS region")
	s3Cmd.Flags().StringP("bucket", "b", "", "S3 bucket")
	s3Cmd.Flags().Bool("role", false, "Use EC2 role for getting credentials")
}

func parseBaseArgs(cmd *cobra.Command) CommandConfig {
	port, _ := cmd.PersistentFlags().GetInt("port")
	cronSpec, _ := cmd.PersistentFlags().GetString("cron")
	domain, _ := cmd.PersistentFlags().GetString("domain")
	ssl, _ := cmd.PersistentFlags().GetBool("ssl")

	return CommandConfig{port, cronSpec, domain, ssl}
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

func serve(cmdConf CommandConfig, hitStorage storage.HitStorage) {
	var mgr *autocert.Manager

	hitServer := server.NewHitServer()
	go hitServer.StartWorker()

	c := cron.New()

	c.AddFunc("*/5 * * * *", func() {
		log.Println("Saving cache to disk")
		err := hitServer.SaveCache()
		if err != nil {
			log.Fatal(err)
		}
	})

	c.AddFunc(cmdConf.cronSpec, func() {
		log.Println("Archiving...")
		err := archiveAndClearCache(&hitServer, hitStorage)
		if err != nil {
			log.Println(err)
		}
	})

	c.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/", hitServer.HandlePixelRequest)
	mux.HandleFunc("/hitpoints.js", hitServer.HandleJS(cmdConf.domain, cmdConf.ssl))
	srv := serverFromMux(mux)

	if cmdConf.ssl {
		mgr = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache("certs"),
			HostPolicy: autocert.HostWhitelist(cmdConf.domain),
		}
		// Spinning up main HTTP port in goroutine
		go func() {
			log.Println("Starting redirect server...")
			redirectSrv := serverFromMux(nil)
			redirectSrv.Addr = fmt.Sprintf(":%d", cmdConf.port)
			redirectSrv.Handler = mgr.HTTPHandler(nil)
			err := redirectSrv.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}()

		srv.Addr = ":443"
		srv.TLSConfig = &tls.Config{GetCertificate: mgr.GetCertificate}

		log.Println("Starting TLS server...")
		err := srv.ListenAndServeTLS("", "")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Starting server...")
		srv.Addr = fmt.Sprintf(":%d", cmdConf.port)
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}
}
