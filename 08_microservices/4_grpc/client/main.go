package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"

	"gitlab.vk-golang.ru/vk-golang/lectures/08_microservices/4_grpc/session"
)

var loginFormTmpl = []byte(`
<html>
	<body>
	<form action="/login" method="post">
		Login: <input type="text" name="login">
		Password: <input type="password" name="password">
		<input type="submit" value="Login">
	</form>
	</body>
</html>
`)

var (
	sessManager session.AuthCheckerClient
)

func checkSession(r *http.Request) (*session.Session, error) {
	cookieSessionID, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	sess, err := sessManager.Check(
		context.Background(),
		&session.SessionID{
			ID: cookieSessionID.Value,
		})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func innerPage(w http.ResponseWriter, r *http.Request) {
	sess, err := checkSession(r)
	if err != nil {
		w.Write(loginFormTmpl)
		return
	}
	if sess == nil {
		w.Write(loginFormTmpl)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, "Welcome, "+sess.Login+" <br />")
	fmt.Fprintln(w, "Session ua: "+sess.Useragent+" <br />")
	fmt.Fprintln(w, `<a href="/logout">logout</a>`)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	inputLogin := r.FormValue("login")
	expiration := time.Now().Add(365 * 24 * time.Hour)

	sess, err := sessManager.Create(
		context.Background(),
		&session.Session{
			Login:     inputLogin,
			Useragent: r.UserAgent(),
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:    "session_id",
		Value:   sess.ID,
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {

	grpcConn, err := grpc.Dial(
		"127.0.0.1:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("cant connect to grpc")
	}
	defer grpcConn.Close()

	sessManager = session.NewAuthCheckerClient(grpcConn)

	http.HandleFunc("/", innerPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logoutPage)
	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	cookieSessionID, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = sessManager.Delete(
		context.Background(),
		&session.SessionID{
			ID: cookieSessionID.Value,
		})

	cookieSessionID.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, cookieSessionID)

	http.Redirect(w, r, "/", http.StatusFound)
}
