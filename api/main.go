package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/guobinqiu/appdeployer/api/controller"
	"github.com/guobinqiu/appdeployer/api/middleware"
)

func main() {
	// db := service.NewAppDeployerDB("test.db")
	// db.Migrate()

	vmDeployer := controller.NewVMDeployer()
	kubeDeployer := controller.NewKubeDeployer()

	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gin.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())
	r.Use(middleware.Default())

	//userController := controller.NewUserController(db.GetDB())
	//r.POST("/login", userController.Login)

	//r.Use(middleware.Auth()) //鉴权层

	r.POST("/kube/submit", kubeDeployer.Submit)
	r.GET("/kube/deploy", kubeDeployer.Deploy)

	r.POST("/vm/submit", vmDeployer.Submit)
	r.GET("/vm/deploy", vmDeployer.Deploy)

	srv := &http.Server{
		Addr:    ":8888",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
