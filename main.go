// @title Shopping List API
// @version 0.1
// @host localhost:8888
// @BasePath /v1
// @accept json
// @produce json
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	// "strconv"
	"strings"
	// "time"
	"golang.org/x/crypto/bcrypt"
	// _ "net/http/pprof"
	"log/slog"
	// "go.opentelemetry.io/otel/codes"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// "go.opentelemetry.io/otel/attribute"
	// "go.opentelemetry.io/otel/trace"
)



const(
	RoleAdmin Role = "admin"
	RoleUser Role = "user"
)



var (
	db RepositoryInterface
	allData = ShoppingList{}
	listsCache *lru.Cache[string, *ShoppingList]
	metricsService *Metrics
)

// var tracer trace.Tracer

func main(){
	var err error

	listsCache, err = lru.New[string, *ShoppingList](128)

	if err != nil{
		fmt.Println("Unable to initialize lists cache:", err.Error())
		os.Exit(1)
	}

	// if err := initTracer(); err != nil{
	// 	fmt.Println("Unable to initialize tracer:", err.Error())
	// }

	// tracer = tp.Tracer("github.com/smikesword/myapi/main")

	metricsService = NewMetrics()
	db, err = NewRepository("./app.db")
	



	if err != nil {
		fmt.Println("Unable to open db:", err.Error())
		os.Exit(1)
	}

	if err := db.Init(); err != nil {
		fmt.Println("Unable to initialize db:", err.Error())
		os.Exit(1)
	}


	// http := http.NewServehttp()

	http.HandleFunc("GET /v1/lists", authRequired(handleGetLists))
	http.HandleFunc("GET /v1/lists/{id}", authRequired(handleGetList))
	http.HandleFunc("POST /v1/lists", MetricsMiddleware(authRequired(handleCreateList)))
	http.HandleFunc("DELETE /v1/lists/{id}", adminRequired(handleDeleteList))
	http.HandleFunc("PUT /v1/lists/{id}", MetricsMiddleware(authRequired(handleUpdateList)))
	http.HandleFunc("PATCH /v1/lists/{id}", MetricsMiddleware(authRequired(handlePatchList)))
	// http.HandleFunc("POST /v1/lists/{id}", authRequired(handleListPush))
	http.HandleFunc("GET /health", handleHealth)
	http.HandleFunc("POST /login", handleLogin)
	fmt.Println("listening on port :8888")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8888", nil)
}


func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //default cost is 10
	return string(bytes), err
}

func validatePassword(password, hash string) bool{
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if !strings.HasPrefix(token, "Bearer "){
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token = token[7:]

		_, err := db.GetSession(token)

		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func adminRequired(next http.HandlerFunc) http.HandlerFunc {
	return authRequired (func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		token = token[7:]
		user, err := db.GetUserFromToken(token)

		if err != nil || user.Role != RoleAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		} 

		next(w, r)
	})
}



func handleHealth(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}


// @Summary Gets all shopping lists
// @Description Retrieve a list of all shopping lists
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} ShoppingList
// @Failure 401 {string} string "unauthorized"
// @Router /lists [get]
func handleGetLists(w http.ResponseWriter, r *http.Request){
	// _, span := tracer.Start(r.Context(), "get-lists")
	// defer span.End()
	// dbSummary := attribute.String("db.query.summary","SELECT * FROM lists")
	// span.SetAttributes(dbSummary)
	lists, err := db.GetShoppingLists()

	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		// span.SetStatus(codes.Error, "Failed to get lists from DB")
		return
	}
	

	data, err := json.Marshal(lists)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// @Summary Gets a specific shopping list by ID
// @Description Retrieve a specific shopping list by its ID
// @Accept json
// @Produce json
// @Param id path int true "Shopping List ID"
// @Security ApiKeyAuth
// @Success 200 {object} ShoppingList
// @Failure 401 {string} string "unauthorized"
// @Failure 404 {string} string "List not found"
// @Router /lists/{id} [get]
func handleGetList(w http.ResponseWriter, r *http.Request){
	
	id := r.PathValue("id")
	list, found := listsCache.Get(id)
	var err error
	if !found {
		list, err = db.GetShoppingList(id)

		if err != nil {
			slog.Info("List not found in database for ID: %s", id, slog.Any("error", err))
			http.Error(w, "List not found", http.StatusNotFound)
			return
		}
		
		listsCache.Add(id, list)
		
	}


	data, err := json.Marshal(list)
		
	if err != nil {
		slog.Error("Failed to marshal list data for ID:", id, slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-cache") //will check if data in cache is still valid

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)

	_, err = w.Write(data)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
}

func handleCreateList(w http.ResponseWriter, r *http.Request){
	// _, span := tracer.Start(r.Context(), "create-list")
	// defer span.End()
	var listDTO ShoppingListDTO

	// httpMethodAttrib := attribute.String("http.request.method", r.Method)
	// httpRouteAttrib := attribute.String("http.request.url", r.URL.Path)

	// span.SetAttributes(httpMethodAttrib)
	// span.SetAttributes(httpRouteAttrib)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Disallow unknown fields to prevent unexpected data
	err := decoder.Decode(&listDTO)
	

	if err != nil{
		slog.Error("Invalid request body", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if listDTO.Name == "" || len(listDTO.Items) == 0 || listDTO.Items == nil {
		slog.Error("Invalid request body: Name and Items are required")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err = db.CreateShoppingList(&listDTO)

	if err != nil{
		slog.Error("Failed to create shopping list", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(listDTO)

	if err != nil{
		slog.Error("Failed to marshal list data", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleDeleteList(w http.ResponseWriter, r *http.Request){
	id := r.PathValue("id")
	err := db.DeleteShoppingList(id)

	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleUpdateList(w http.ResponseWriter, r *http.Request){
	id := r.PathValue("id")
	var listDTO ShoppingListDTO


	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&listDTO)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.UpdateShoppingList(id, &listDTO)

	if err != nil{
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	listsCache.Remove(id)

	if err := json.NewEncoder(w).Encode(listDTO); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
}

func handlePatchList(w http.ResponseWriter, r *http.Request){
	id := r.PathValue("id")
	var listPatch ShoppingListPatch

	err := json.NewDecoder(r.Body).Decode(&listPatch)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = db.PatchShoppingList(id, &listPatch)

	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
	}

	listsCache.Remove(id)

	err = json.NewEncoder(w).Encode(listPatch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
}

// func handleListPush(w http.ResponseWriter, r *http.Request){
// 	id := r.PathValue("id")
// 	for i, list := range allData {
// 		if strconv.Itoa(list.ID) == id {
// 		var item ListPushAction
// 		err := json.NewDecoder(r.Body).Decode(&item)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		list.Items = append(list.Items, item.Item)
// 		allData[i] = list
// 		err = json.NewEncoder(w).Encode(list)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		return
// 		}
// 	}
// 	http.Error(w, "List not found", http.StatusNotFound)
// }

func handleLogin(w http.ResponseWriter, r *http.Request){
	var loginData LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := db.GetUser(loginData)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session, err := db.AddSession(user.Username)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": session.Token})

}