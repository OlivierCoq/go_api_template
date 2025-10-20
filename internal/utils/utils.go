package utils

import (
	"encoding/json" //Marhsaling and Unmarshaling JSON come from this package
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Envelope is a generic map for wrapping JSON responses. It allows for flexible response structures. Kind of like Typescript's Record<string, any>
type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Envelope) {

	// MarshalIndent is used to convert the data into a pretty-printed JSON format.
	// The second parameter is the prefix (empty string means no prefix),
	// and the third parameter is the indentation (two spaces in this case).
	js, err := json.MarshalIndent(data, "", "  ") // how many tabs/spaces for indentation
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	js = append(js, '\n') // add a newline at the end for better readability in the console
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}

func ReadIDParam(r *http.Request, param string) (int64, error) {
	idParam := chi.URLParam(r, param)
	if idParam == "" {
		return 0, fmt.Errorf("missing %s parameter", param)
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", param)
	}
	return id, nil
}
