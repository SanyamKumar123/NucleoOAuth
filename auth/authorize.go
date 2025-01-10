package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type AuthorizeRoute struct {
	Rdb *redis.Client
	Db  *sql.DB
}

func (s *AuthorizeRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Getting Return URL
	returnURL := r.URL.Query()["returnURL"]
	if len(returnURL) == 0 {
		io.WriteString(w, "ERROR: No ReturnUrl provided in Query Params")
		return
	}

	//Generating a random ID for current session
	id := generateRandomID(s.Rdb)

	//Declaring the Current Session and Converting to String
	curr_session := Session{ReturnUrl: returnURL[0], Id: id}
	data, _ := json.Marshal(curr_session)
	session_str := string(data)

	//Writing to Cache the ID and returnURL
	s.Rdb.Set(context.Background(), id, session_str, ValidTime)

	//Opening a new window to authenticate user
	code := fmt.Sprintf("<script>window.open('http://localhost:5173/%s', 'targetWindow', 'menubar=1,resizeable=1,width=500,height=600');</script>", id)
	io.WriteString(w, code)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateRandomID(rdb *redis.Client) string {
	b := make([]rune, 16)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	id := string(b)

	_, err := rdb.Get(context.Background(), id).Result()
	if err != redis.Nil {
		return generateRandomID(rdb)
	}

	return id
}
