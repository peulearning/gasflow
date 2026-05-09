package gateway

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"github.com/peulearning/gasflow/api/internal/infra/auth"
	analyticsH "github.com/peulearning/gasflow/api/internal/modules/analytics"
	billingH "github.com/peulearning/gasflow/api/internal/modules/billing"
	clientsH "github.com/peulearning/gasflow/api/internal/modules/clients"
	inventoryH "github.com/peulearning/gasflow/api/internal/modules/inventory"
	ordersH "github.com/peulearning/gasflow/api/internal/modules/orders"
	"golang.org/x/crypto/bcrypt"
)

// Handlers agrupa todos os handlers de módulos.
type Handlers struct {
	Clients   *clientsH.Handler
	Orders    *ordersH.Handler
	Inventory *inventoryH.Handler
	Billing   *billingH.Handler
	Analytics *analyticsH.Handler
	Auth      *auth.Service
	DB        *sql.DB // necessário para o login handler
}

// New monta e retorna o router HTTP completo.
func New(h Handlers, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()

	// ── Middlewares globais ────────────────────────────────────────────────
	r.Use(chimid.RequestID)
	r.Use(chimid.RealIP)
	r.Use(Recoverer)
	r.Use(Logger)
	r.Use(CORS(allowedOrigins))

	// ── Health check (sem auth) ────────────────────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		OK(w, map[string]string{"status": "ok"})
	})

	// ── Auth (sem autenticação) ────────────────────────────────────────────
	r.Post("/api/auth/login", loginHandler(h.Auth, h.DB))

	// ── Rotas protegidas por JWT ───────────────────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(h.Auth.Authenticate)

		// Clientes — admin e operacional
		r.Route("/api/clients", func(r chi.Router) {
			h.Clients.Routes(r)
		})

		// Pedidos — admin e operacional
		r.Route("/api/orders", func(r chi.Router) {
			h.Orders.Routes(r)
		})

		// Estoque — admin e operacional
		r.Route("/api/inventory", func(r chi.Router) {
			h.Inventory.Routes(r)
		})

		// Cobranças — admin e financeiro apenas
		r.Route("/api/charges", func(r chi.Router) {
			r.Use(h.Auth.Authorize(auth.RoleAdmin, auth.RoleFinancial))
			h.Billing.Routes(r)
		})

		// Dashboard — todos os roles autenticados
		r.Route("/api/dashboard", func(r chi.Router) {
			h.Analytics.Routes(r)
		})
	})

	return r
}

// ── Login handler ─────────────────────────────────────────────────────────────

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
}

func loginHandler(svc *auth.Service, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			BadRequest(w, "payload inválido")
			return
		}
		if req.Email == "" || req.Password == "" {
			BadRequest(w, "email e password são obrigatórios")
			return
		}

		var userID, name, hash, role string
		err := db.QueryRowContext(r.Context(),
			`SELECT id, name, password, role FROM users WHERE email=? AND is_active=1`,
			req.Email,
		).Scan(&userID, &name, &hash, &role)
		if err != nil {
			Unauthorized(w, "credenciais inválidas")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			Unauthorized(w, "credenciais inválidas")
			return
		}

		access, err := svc.GenerateAccessToken(userID, req.Email, auth.Role(role))
		if err != nil {
			InternalError(w, "erro ao gerar token")
			return
		}
		refresh, _ := svc.GenerateRefreshToken(userID)

		OK(w, loginResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			UserID:       userID,
			Role:         role,
		})
	}
}