package main

import (
	"avitoMerchStore/internal/connections"
	"avitoMerchStore/internal/handler"
	"avitoMerchStore/internal/repository"
	"avitoMerchStore/internal/service"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	conn := connections.InitPool(ctx)
	userRepository := repository.NewUserRepository(conn)
	userService := service.NewUserService(userRepository)
	hu := handler.NewUserHandler(userService)

	teamRepository := repository.NewTeamRepository(conn)
	teamService := service.NewTeamService(teamRepository)
	ht := handler.NewTeamHandler(teamService)

	prRepository := repository.NewPrRepository(conn)
	prService := service.NewPrService(prRepository)
	hp := handler.NewPrHandler(prService)

	r := chi.NewRouter()

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", hu.SetIsActive)
		r.Get("/getReview", hu.GetPrById)
	})
	r.Route("/team", func(r chi.Router) {
		r.Post("/add", ht.AddTeam)
		r.Get("/get", ht.GetTeam)
	})
	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", hp.CreatePR)
		r.Post("/merge", hp.MergePR)
		r.Post("/reassign", hp.ReassignPR)
	})

	r.Head("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	pflag.StringVar(&port, "port", port, "The port to listen on")
	pflag.Parse()
	fmt.Println("Listening on port " + port)
	srv := &http.Server{Addr: ":" + port, Handler: r}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		err = srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(shutdownCtx)
	log.Println("Shutting down service-courier")
	defer conn.Close()
}
