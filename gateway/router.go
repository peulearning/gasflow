package gateway

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"

	"gasflow/internal/httputil"
	"gasflow/infra/auth"
	analyticsH "gasflow/modules/analytics"
	billingH   "gasflow/modules/billing"
	clientsH   "gasflow/modules/clients"
	inventoryH "gasflow/modules/inventory"
	ordersH    "gasflow/modules/orders"
)

// Handlers agrupa todos os handlers registrados no router.
type Handlers struct {
	Clients   *clientsH.Handler
	Orders    *ordersH.Handler
	Inventory *inventoryH.Handler
	Billing   *billingH.Handler
	Analytics *analyticsH.Handler
	Auth      *auth.Service
	DB        *sql.DB
}

// New constrói e retorna o router HTTP completo.
func New(h Handlers, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()

	r.Use(chimid.RequestID)
	r.Use(chimid.RealIP)
	r.Use(Recoverer)
	r.Use(Logger)
	r.Use(CORS(allowedOrigins))

	// Sem autenticação
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.OK(w, map[string]string{"status": "ok", "service": "gasflow"})
	})
	r.Post("/api/auth/login", loginHandler(h.Auth, h.DB))

	// Rotas protegidas
	r.Group(func(r chi.Router) {
		r.Use(h.Auth.Authenticate)

		r.Route("/api/clients", func(r chi.Router) {
			h.Clients.Routes(r)
		})
		r.Route("/api/orders", func(r chi.Router) {
			h.Orders.Routes(r)
		})
		r.Route("/api/inventory", func(r chi.Router) {
			h.Inventory.Routes(r)
		})
		r.Route("/api/charges", func(r chi.Router) {
			r.Use(h.Auth.Authorize(auth.RoleAdmin, auth.RoleFinancial))
			h.Billing.Routes(r)
		})
		r.Route("/api/dashboard", func(r chi.Router) {
			h.Analytics.Routes(r)
		})
	})

	return r
}

// ── Login handler ─────────────────────────────────────────────────────────────

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
}

func loginHandler(svc *auth.Service, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.BadRequest(w, "payload inválido")
			return
		}
		if req.Email == "" || req.Password == "" {
			httputil.BadRequest(w, "email e password obrigatórios")
			return
		}

		var userID, hash, role string
		err := db.QueryRowContext(r.Context(),
			`SELECT id, password, role FROM users WHERE email=? AND is_active=1`,
			req.Email,
		).Scan(&userID, &hash, &role)
		if err != nil {
			// Tempo constante para evitar timing attack
			bcrypt.CompareHashAndPassword([]byte("$2a$10$dummydummydummydummyduu"), []byte(req.Password))
			httputil.Unauthorized(w, "credenciais inválidas")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			httputil.Unauthorized(w, "credenciais inválidas")
			return
		}

		access, err := svc.GenerateAccessToken(userID, req.Email, auth.Role(role))
		if err != nil {
			httputil.InternalError(w, "erro ao gerar token")
			return
		}
		refresh, _ := svc.GenerateRefreshToken(userID)

		httputil.OK(w, loginResp{
			AccessToken:  access,
			RefreshToken: refresh,
			UserID:       userID,
			Role:         role,
		})
	}
}