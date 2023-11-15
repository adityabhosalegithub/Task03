package function

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

type Employee struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"firstname"`
	LastName  string  `json:"lastname"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	Role      string  `json:"role"`
	Salary    float64 `json:"salary"`
	PhoneNo   string  `json:"phone"`
}

const (
	projectId      string = "gifted-pulsar-402404"
	collectionName string = "employees"
)

func init() {
	functions.HTTP("AddEmployee", addEmployee)
}

func addEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var newEmployee Employee

	err := json.NewDecoder(r.Body).Decode(&newEmployee)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validations := []struct {
		fieldName string
		check     func(string) error
		value     string
	}{
		{"First Name", IsValidName, newEmployee.FirstName},
		{"Last Name", IsValidName, newEmployee.LastName},
		{"Email", IsValidEmail, newEmployee.Email},
		{"Role", IsValidRole, newEmployee.Role},
		{"Salary", IsValidSalary, fmt.Sprint(newEmployee.Salary)},
		{"Phone Number", IsValidPhoneNo, newEmployee.PhoneNo},
	}

	for _, validation := range validations {
		if err := validation.check(validation.value); err != nil {
			http.Error(w, fmt.Sprintf("%s: %s", validation.fieldName, err.Error()), http.StatusBadRequest)
			return
		}
	}

	if err := IsIdUnique(newEmployee.ID); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		http.Error(w, "Failed to create Firestore Client", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	employeesCollection := client.Collection(collectionName)
	_, _, err = employeesCollection.Add(ctx, map[string]interface{}{
		"id":        newEmployee.ID,
		"firstname": newEmployee.FirstName,
		"lastname":  newEmployee.LastName,
		"email":     newEmployee.Email,
		"password":  newEmployee.Password,
		"role":      newEmployee.Role,
		"salary":    newEmployee.Salary,
		"phone":     newEmployee.PhoneNo,
	})

	if err != nil {
		http.Error(w, "Failed to add employee", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Employee added successfully")
}

func IsValidName(value string) error {
	if value == "" {
		return fmt.Errorf("cannot be empty. Enter again")
	}
	return nil
}

func IsValidEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.]+@[a-zA-Z0-9]+.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("email should contain @ and . Enter valid email")
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

func IsValidRole(role string) error {
	role = strings.ToLower(role)
	if role != "admin" && role != "manager" && role != "developer" && role != "tester" {
		return fmt.Errorf("given role cannot be accepted")
	}
	return nil
}

func IsValidSalary(salaryStr string) error {
	salary, err := strconv.ParseFloat(salaryStr, 64)
	if err != nil || salary == 0 {
		return fmt.Errorf("salary cannot be ZERO. Enter salary again")
	}
	return nil
}

func IsIdUnique(id int64) error {
	ctx := context.Background()
	client, _ := firestore.NewClient(ctx, projectId)
	defer client.Close()

	collectionRef := client.Collection(collectionName)
	query := collectionRef.Where("ID", "==", id).Limit(1)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("error checking if ID is unique")
	}

	if len(docs) > 0 {
		return fmt.Errorf("ID already in use. Provide a unique ID")
	}

	return nil
}
