package apihandler

import (
	"net/http"
)

func (ah APIHandler) SignupNewCustomer(w http.ResponseWriter, r *http.Request) {
	// token := r.Header.Get("Authorization")

	// signupRequest := server.SignupRequest{}
	// if ok := ah.ReadRequestBody(w, r, &signupRequest); !ok {
	// 	return
	// }
	// // map to domain model
	// domainSignup := mapSignupToDomain(&signupRequest)

	// // call usecase
	// err := ah.AppLogic.Signup(r.Context(), domainSignup, token)
	// if err != nil {
	// 	ah.HandleErrors(w, r, err, "Signup failed")
	// 	return
	// }

	// return success
	w.WriteHeader(http.StatusCreated)
}
