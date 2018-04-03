package main

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/companies", a.getCompanies).Methods("GET")
	a.Router.HandleFunc("/company/{id:[0-9]+}", a.getCompany).Methods("GET")
}
