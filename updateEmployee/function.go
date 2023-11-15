package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/iterator"
)

var client *firestore.Client

const (
	projectId      string = "gifted-pulsar-402404"
	collectionName string = "employees"
)

func init() {
	ctx := context.Background()
	var err error
	client, err = firestore.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("Failed to create Firestore Client: %v", err)
	}
	functions.HTTP("UpdateEmployee", updateEmployee)
}

type Employee struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"firstname"`
	LastName  string  `json:"lastname"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	PhoneNo   string  `json:"phone"`
	Role      string  `json:"role"`
	Salary    float64 `json:"salary"`
}

func updateEmployee(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	idParam := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	if !IsEmpPresent(id) {
		http.Error(w, "Employee with given id not found", http.StatusNotFound)
		return
	}

	var updatedEmployee Employee
	err = json.NewDecoder(r.Body).Decode(&updatedEmployee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedEmployee.FirstName = formatName(updatedEmployee.FirstName)
	updatedEmployee.LastName = formatName(updatedEmployee.LastName)
	updatedEmployee.Role = strings.ToLower(updatedEmployee.Role)

	if !validateAllUpdatingDatafield(id, w, updatedEmployee) {
		return
	}

	employeesCollection := client.Collection(collectionName)

	iter := employeesCollection.Where("ID", "==", id).Documents(ctx)
	d, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			http.Error(w, fmt.Sprintf("Employee not found with ID: %d", id), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to query Firestore: %v", err), http.StatusInternalServerError)
		return
	}

	_, err = d.Ref.Update(ctx, []firestore.Update{
		{Path: "ID", Value: updatedEmployee.ID},
		{Path: "FirstName", Value: updatedEmployee.FirstName},
		{Path: "LastName", Value: updatedEmployee.LastName},
		{Path: "Email", Value: updatedEmployee.Email},
		{Path: "Password", Value: updatedEmployee.Password},
		{Path: "PhoneNo", Value: updatedEmployee.PhoneNo},
		{Path: "Role", Value: updatedEmployee.Role},
		{Path: "Salary", Value: updatedEmployee.Salary},
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update employee: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Employee with ID %d has been updated successfully", id)
}

func validateAllUpdatingDatafield(id int, w http.ResponseWriter, newEmployee Employee) bool {

	// Validate First Name
	firstname := newEmployee.FirstName
	firstname = strings.TrimSpace(firstname)
	if err := IsValidName(firstname); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// Validate Last Name
	lastname := newEmployee.LastName
	lastname = strings.TrimSpace(lastname)
	if err := IsValidName(lastname); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// Validate Email
	email := newEmployee.Email
	email = strings.TrimSpace(email)
	if err := IsValidEmail(email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// Validate Phone Number
	phoneno := newEmployee.PhoneNo
	phoneno = strings.TrimSpace(phoneno)
	err := IsValidPhoneNo(phoneno)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// Validate Role
	role := newEmployee.Role
	role = strings.TrimSpace(role)
	if err := IsValidRole(role); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	// Validate Salary
	salary := newEmployee.Salary
	if err := IsValidSalary(salary); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func IsEmpPresent(id int) bool {
	ctx := context.Background()
	collectionRef := client.Collection(collectionName)
	query := collectionRef.Where("ID", "==", id).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error querying Firestore: %v", err)
		return false
	}
	return len(docs) > 0
}

func IsValidName(value string) error {
	if value == "" {
		return fmt.Errorf("cannot be empty. Enter again")
	}
	return nil
}

func IsValidRole(role string) error {
	role = strings.ToLower(role)
	if role != "admin" && role != "manager" && role != "developer" && role != "tester" {
		return fmt.Errorf("given role cannot be accepted")
	}
	return nil
}

func IsValidSalary(salaryStr float64) error {
	// salary, err := strconv.ParseFloat(salaryStr, 64)
	if salaryStr == 0 {
		return fmt.Errorf("salary cannot be ZERO. Enter salary again")
	}
	return nil
}

func IsValidPhoneNo(phoneNo string) error {
	if len(phoneNo) != 10 {
		return fmt.Errorf("phone number must be 10 digits long")
	}
	_, err := strconv.ParseInt(phoneNo, 10, 64)
	if err != nil {
		return fmt.Errorf("gailed to convert phone number to integer")
	}
	return nil
}
func formatName(name string) string {
	return strings.ToUpper(string(name[0])) + strings.ToLower(strings.TrimSpace(name[1:]))
}

func IsValidEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.]+@[a-zA-Z0-9]+.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("email should contain @ and . Enter valid email")
	}
	return nil
}
