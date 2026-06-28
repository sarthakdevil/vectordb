package main
import (
	"encoding/json"
	"fmt"
	"net/http"
)
type HomeMessage struct {
	Text string `json:"text"`
}

func runningmsg(w http.ResponseWriter, r *http.Request){
	msg := HomeMessage{
		Text: "hello message",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}
func main() {
	http.HandleFunc("/", runningmsg)
	http.HandleFunc("/insert",insertvectors)
	http.HandleFunc("/delete",Deletevectors)
	http.HandleFunc("/search",handleSearch)
	http.HandleFunc("/create-index",createindex)
	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}