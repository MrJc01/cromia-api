package middleware

import (
	"cromia/api/internal/db"
	"cromia/api/internal/utils"
	"net/http"
)

func BillingMiddleware(database db.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInter := r.Context().Value(UserContextKey)
		if userInter == nil {
			utils.JSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		cachedUser := userInter.(*db.User)

		freshUser, err := database.GetUserByID(cachedUser.ID)
		if err != nil || freshUser == nil {
			utils.JSONError(w, "user not found", http.StatusUnauthorized)
			return
		}

		if freshUser.Balance <= 0 {
			utils.JSONError(w, "Insufficient credits. Please top up your account.", http.StatusPaymentRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}
