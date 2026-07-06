package main

import (
	"log"
	"net/http"

	"latrode-fusion/internal/config"
	"latrode-fusion/internal/database"
	"latrode-fusion/internal/handlers"
	"latrode-fusion/internal/middleware"
	"latrode-fusion/internal/repository"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepo(db)
	cartRepo := repository.NewCartRepo(db)
	favRepo := repository.NewFavoriteRepo(db)
	orderRepo := repository.NewOrderRepo(db)

	resetRepo := repository.NewPasswordResetRepo(db)
	authHandler := handlers.NewAuthHandler(userRepo, resetRepo, cfg)
	productHandler := handlers.NewProductHandler(productRepo)
	cartHandler := handlers.NewCartHandler(cartRepo)
	favHandler := handlers.NewFavoriteHandler(favRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, cartRepo, cfg)
	adminHandler := handlers.NewAdminHandler(orderRepo, productRepo, userRepo)
	paymentHandler := handlers.NewPaymentHandler(orderRepo)
	oauthHandler := handlers.NewOAuthHandler(cfg, userRepo)

	auth := middleware.Auth(userRepo)
	adminEmail := "latrode.co@gmail.com"
	adminAuth := middleware.AdminOnly(adminEmail)
	optionalAuth := middleware.OptionalAuth(userRepo)

	wrap := func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(fn)
	}

	api := http.NewServeMux()

	api.HandleFunc("GET /products", productHandler.List)
	api.HandleFunc("GET /products/{id}", productHandler.Get)
	api.HandleFunc("GET /payment-methods", paymentHandler.ListMethods)

	api.HandleFunc("GET /auth/google/login", oauthHandler.GoogleLogin)
	api.HandleFunc("GET /auth/google/callback", oauthHandler.GoogleCallback)

	api.HandleFunc("POST /auth/register", authHandler.Register)
	api.HandleFunc("POST /auth/login", authHandler.Login)
	api.HandleFunc("POST /auth/logout", authHandler.Logout)
	api.HandleFunc("POST /auth/forgot-password", authHandler.ForgotPassword)
	api.HandleFunc("POST /auth/verify-reset-code", authHandler.VerifyResetCode)
	api.HandleFunc("POST /auth/reset-password", authHandler.ResetPassword)
	api.Handle("GET /auth/profile", auth(wrap(authHandler.GetProfile)))
	api.Handle("PUT /auth/profile", auth(wrap(authHandler.UpdateProfile)))
	api.Handle("POST /auth/set-password", auth(wrap(authHandler.SetPassword)))

	api.Handle("GET /cart", optionalAuth(wrap(cartHandler.GetCart)))
	api.Handle("POST /cart", optionalAuth(wrap(cartHandler.AddToCart)))
	api.Handle("PUT /cart/{id}", optionalAuth(wrap(cartHandler.UpdateCartItem)))
	api.Handle("DELETE /cart/{id}", optionalAuth(wrap(cartHandler.RemoveFromCart)))

	api.Handle("GET /favorites", optionalAuth(wrap(favHandler.List)))
	api.Handle("POST /favorites", optionalAuth(wrap(favHandler.Add)))
	api.Handle("DELETE /favorites/{id}", optionalAuth(wrap(favHandler.Remove)))

	api.Handle("POST /orders", optionalAuth(wrap(orderHandler.Create)))
	api.Handle("GET /orders/my", optionalAuth(wrap(orderHandler.GetMyOrders)))
	api.Handle("GET /orders/{id}", optionalAuth(wrap(orderHandler.Get)))

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /dashboard/stats", adminHandler.Dashboard)
	adminMux.HandleFunc("GET /orders", adminHandler.ListOrders)
	adminMux.HandleFunc("PUT /orders/{id}/status", adminHandler.UpdateOrderStatus)
	adminMux.HandleFunc("GET /products", adminHandler.ListProducts)
	adminMux.HandleFunc("POST /products", adminHandler.CreateProduct)
	adminMux.HandleFunc("PUT /products/{id}", adminHandler.UpdateProduct)
	adminMux.HandleFunc("DELETE /products/{id}", adminHandler.DeleteProduct)
	adminMux.HandleFunc("GET /users", adminHandler.ListUsers)
	adminMux.HandleFunc("GET /payment-methods", adminHandler.GetPaymentMethods)
	adminMux.HandleFunc("GET /settings", adminHandler.GetSettings)
	adminMux.HandleFunc("PUT /settings", adminHandler.UpdateSetting)
	adminMux.HandleFunc("GET /logs", adminHandler.GetLogs)
	api.Handle("/admin/", auth(adminAuth(http.StripPrefix("/admin", adminMux))))

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	fs := http.FileServer(http.Dir(cfg.Frontend))
	mux.Handle("/assets/", fs)
	mux.Handle("/css/", fs)
	mux.Handle("/js/", fs)
	mux.Handle("/admin/", fs)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, cfg.Frontend+"/index.html")
	})

	handler := middleware.CORS(mux)

	log.Printf("Servidor iniciado en http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
