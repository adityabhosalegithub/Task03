package function

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/iterator"
)

func init() {
	functions.HTTP("ViewAllEmployee", viewAllEmployee)
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

const (
	projectId      string = "gifted-pulsar-402404"
	collectionName string = "employees"
)

func viewAllEmployee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Failed to create Firestore Client")
		return
	}
	defer client.Close()

	employeesCollection := client.Collection(collectionName)

	iter := employeesCollection.Documents(ctx)

	var employees []Employee

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode("Failed to retrieve employees")
			return
		}

		var emp Employee
		doc.DataTo(&emp)
		employees = append(employees, emp)
	}

	json.NewEncoder(w).Encode(employees)
}
