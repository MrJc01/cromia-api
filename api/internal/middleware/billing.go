package middleware

import (
	"cromia/api/internal/db"
	"net/http"
)

func BillingMiddleware(database db.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInter := r.Context().Value(UserContextKey)
		if userInter == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		cachedUser := userInter.(*db.User)

		freshUser, err := database.GetUserByID(cachedUser.ID)
		if err != nil || freshUser == nil {
			http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
			return
		}

		if freshUser.Balance <= 0 {
			http.Error(w, `{"error":"Insufficient credits. Please top up your account."}`, http.StatusPaymentRequired)
			return
		}

		next.ServeHTTP(w, r)
	})
}
