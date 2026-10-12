package main

import (
	"database/sql"
	"math/rand"
	"strconv"
	"time"
	sq "github.com/Masterminds/squirrel"
	_ "modernc.org/sqlite"
	"strings"
)

type RepositoryInterface interface {
	Init() error
	AddSession(username string) (*Session, error)
	GetSession(token string) (*Session, error)
	CreateShoppingList(shoppingList *ShoppingListDTO) error
	GetShoppingLists()([]ShoppingList, error)
	GetShoppingList(id string) (*ShoppingList, error)
	UpdateShoppingList(id string, listDTO *ShoppingListDTO) (error)
	PatchShoppingList(id string, patch *ShoppingListPatch) error
	DeleteShoppingList(id string) error
	GetUserFromToken(token string) (*UserDTO, error)
	GetUser(credentials LoginRequest) (*UserDTO, error)
}

type Repository struct {
	db *sql.DB
}

func NewRepository(database string) (*Repository, error){
	db, err := sql.Open("sqlite", database)

	if err != nil {
		return nil, err
	}

	return &Repository{db}, nil
}

func (r *Repository) Init() error {
	if _, err := r.db.Exec("CREATE TABLE IF NOT EXISTS users (role VARCHAR, username VARCHAR PRIMARY KEY, password VARCHAR)"); err != nil {
		return err
	}

	if _, err := r.db.Exec("CREATE TABLE IF NOT EXISTS sessions (token VARCHAR PRIMARY KEY, expires TIMESTAMP, username VARCHAR)"); err != nil {
		return err
	}
	if _, err := r.db.Exec("CREATE TABLE IF NOT EXISTS shopping_lists (id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR, items TEXT)"); err != nil {
		return err
	}

	if _, err := r.db.Exec("INSERT OR IGNORE INTO users (role, username, password) VALUES ('admin', 'admin', 'admin')"); err != nil {
		return err
	}

	if _, err := r.db.Exec("INSERT OR IGNORE INTO users (role, username, password) VALUES ('user', 'user', 'user')"); err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteShoppingList(id string)(error){
	query := sq.Delete("shopping_lists").Where(sq.Eq{"id": id})
	_, err := query.RunWith(r.db).Exec()
	
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) CreateShoppingList(shoppingList *ShoppingListDTO) error{
	query := sq.Insert("shopping_lists").Columns("name", "items").Values(shoppingList.Name, strings.Join(shoppingList.Items, ","))
	_, err := query.RunWith(r.db).Exec()
	return err
}

func (r *Repository) GetShoppingLists()([]ShoppingList, error){
	query := sq.Select("*").From("shopping_lists") 
	rows, err := query.RunWith(r.db).Query()

	if err != nil{
		return nil, err
	}

	lists := []ShoppingList{}

	defer rows.Close()

	for rows.Next(){
		var list ShoppingList
		var items string
		if err := rows.Scan(&list.ID, &list.Name, &items); err != nil {
			return nil, err
		}

		if items != "" {
			list.Items = strings.Split(items, ",")
		}

		lists = append(lists, list)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return lists, nil
}

func (r *Repository) GetShoppingList(id string) (*ShoppingList, error){
	query := sq.Select("*").From("shopping_lists").Where(sq.Eq{"id": id})
	row := query.RunWith(r.db).QueryRow()
	list := ShoppingList{}

	var items string //items is a string separated by comma
	if err := row.Scan(&list.ID, &list.Name, &items); err != nil{
		return nil, err
	}

	if items != ""{
		list.Items = strings.Split(items,",")
	}

	return &list, nil
}

func (r *Repository) UpdateShoppingList(id string, listDTO *ShoppingListDTO) (error){
	query := sq.Update("shopping_lists").Where(sq.Eq{"id": id}).Set("name", listDTO.Name).Set("items", strings.Join(listDTO.Items,","))
	_, err := query.RunWith(r.db).Exec()
	return err
}

func(r *Repository) PatchShoppingList(id string, patch *ShoppingListPatch) error{
	query := sq.Update("shopping_lists").Where(sq.Eq{"id": id})
	
	if patch.Name != nil {
		query = query.Set("name", patch.Name)
	}

	if patch.Items != nil {
		query = query.Set("items", strings.Join(patch.Items, ","))
	}

	_, err := query.RunWith(r.db).Exec()
	if err != nil {
		return err
	}
	return nil

}

func (r *Repository) AddSession(username string) (*Session, error){
	token := strconv.Itoa(rand.Intn(100000000))
	session := Session{Token: token, Expires: time.Now().Add(1 * time.Hour), Username: username }
	query := sq.Insert("sessions").Columns("token", "expires","username").Values(session.Token, session.Expires, session.Username)
	_, err := query.RunWith(r.db).Exec()
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) GetSession(token string) (*Session, error){
	query := sq.Select("token", "expires","username").From("sessions").Where(sq.Eq{"token": token}, sq.Gt{"expires": time.Now()})
	row := query.RunWith(r.db).QueryRow()
	session := Session{}

	if err := row.Scan(&session.Token, &session.Expires, &session.Username); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *Repository) GetUserFromToken(token string) (*UserDTO, error){
	query := sq.Select("role", "username").From("users").Join("sessions USING (username)").Where(sq.Eq{"token":token})
	row := query.RunWith(r.db).QueryRow()

	userDTO := UserDTO{}

	if err := row.Scan(&userDTO.Role, &userDTO.Username); err != nil{
		return nil, err
	}

	return &userDTO, nil
}

func (r *Repository) GetUser(credentials LoginRequest) (*UserDTO, error){
	query := sq.Select("role", "username").From("users").Where(sq.Eq{"username": credentials.Username}, sq.Eq{"password": credentials.Password})
	row := query.RunWith(r.db).QueryRow()

	userDTO := UserDTO{}

	if err := row.Scan(&userDTO.Role, &userDTO.Username); err != nil{
		return nil, err
	}

	return &userDTO, nil
}

// func (r *Repository) PatchShoppingList(id string, patch *ShoppingListPatch) error {
// 	query := sq.Update("shopping_lists").Where(sq.Eq{"id":id})

// 	if patch.Name != nil {
// 		query = query.Set("name", *patch.Name)
// 	}

// 	if patch.Items != nil {
// 		query = query.Set("items", strings.Join(patch.Items, ","))
// 	}

// 	_, err := query.RunWith(r.db).Exec()

// 	if err != nil{
// 		return err
// 	}

// 	return nil
// }