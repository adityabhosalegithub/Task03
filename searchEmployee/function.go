package function

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/iterator"
)

func init() {
	functions.HTTP("SearchEmployee", searchEmployee)
}

type Employee struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"firstname"`
	LastName  string  `json:"lastname"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	Role      string  `json:"role"`
	Salary    float64 `json:"salary"`
}

const (
	projectId      string = "gifted-pulsar-402404"
	collectionName string = "employees"
)

func searchEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the query parameters from the URL
	id := r.URL.Query().Get("id")
	firstName := r.URL.Query().Get("firstName")
	lastName := r.URL.Query().Get("lastName")
	email := r.URL.Query().Get("email")
	role := r.URL.Query().Get("role")

	// Check if at least one search parameter is provided
	if id == "" && firstName == "" && lastName == "" && email == "" && role == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Missing search query parameter")
		return
	}

	// Convert the query parameters to lower case for case-insensitive search

	// Search employees by ID, name, email, or role
	employees, err := SearchEmployees(id, firstName, lastName, email, role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to retrieve employee details")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employees)
}

func SearchEmployees(id, firstName, lastName, email, role string) ([]*Employee, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	employeesCollection := client.Collection(collectionName)

	iter := employeesCollection.Documents(ctx)
	var employees []*Employee

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var employee Employee
		doc.DataTo(&employee)

		// Check if any of the query parameters match any of the employee details
		if (id == "" || strconv.FormatInt(employee.ID, 10) == id) &&
			(firstName == "" || strings.Contains(strings.ToLower(employee.FirstName), strings.ToLower(firstName))) &&
			(lastName == "" || strings.Contains(strings.ToLower(employee.LastName), strings.ToLower(lastName))) &&
			(email == "" || strings.Contains(strings.ToLower(employee.Email), strings.ToLower(email))) &&
			(role == "" || strings.Contains(strings.ToLower(employee.Role), strings.ToLower(role))) {

			employees = append(employees, &employee)
		}
	}

	return employees, nil
}
