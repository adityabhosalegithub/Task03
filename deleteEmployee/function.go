package function

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"cloud.google.com/go/firestore"                                   //go get -u cloud.google.com/go/firestore
	"github.com/GoogleCloudPlatform/functions-framework-go/functions" //go get -u github.com/GoogleCloudPlatform/functions-framework-go@v1.2.0

	//framework for building and deploying Cloud Functions in Go //convert an incoming HTTP request into a function call and to handle the response.
	"google.golang.org/api/iterator" //go get -u google.golang.org/api //package is essential when working with Google APIs that return large sets of data
)

func init() {
	functions.HTTP("DeleteEmployee", deleteEmployee)
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

// declaring Constants
const (
	projectId      string = "gifted-pulsar-402404"
	collectionName string = "employees"
)

// Func to Delete the Employee
func deleteEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//log.Write("Request To Delete Employee")
	// Parse the query string
	employeeIDStr := r.URL.Query().Get("id")
	//if employee id is empty then will show the message
	if employeeIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Missing employee id in the query string")
		//log.Write("Employee Id is not provided")
		return
	}

	// Converting the employeeIDStr to int64
	employeeID, err := strconv.ParseInt(employeeIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid employee id in the query string")
		//log.Write("Invalid Emp Id")
		return
	}

	// Calling the function to delete the employee using Firestore
	msg, err := DeleteEmp(employeeID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(msg)
		//log.Write("Failed To Delete Emp")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(msg)
	//log.Write("Employee Deleted Succesfully")
}

// To delete Employee
func DeleteEmp(id int64) (string, error) {
	ctx := context.Background()
	//creating new client on firestore
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		//log.Write("Failed to create Firestore Client")
		return "Failed To delete Employee", err
	}
	defer client.Close()

	employeesCollection := client.Collection(collectionName)

	// Query Firestore to find the employee with the given email.
	iter := employeesCollection.Where("id", "==", id).Documents(ctx)
	d, err := iter.Next()

	if err != nil {
		// Handle errors, such as when the employee is not found.
		if err == iterator.Done {
			return "employee not found", nil // Employee not found
		}
		return "", err
	}
	d.Ref.Delete(ctx)
	return "Employee Deleted Successfully", nil

}
