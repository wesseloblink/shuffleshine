package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/zmb3/spotify/v2"
)

const redirectURI = "http://127.0.0.1:8080/callback"

var (
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
		),
	)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {
	log.Println(os.Getenv("SPOTIFY_ID"))
	// first start an HTTP server
	http.HandleFunc("/callback", completeAuth)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// err := http.ListenAndServe(":8080", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)
	fmt.Println("Please continue at https://127.0.0.1:8080")
	currentlyPlaying, err := client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		log.Fatal(err)
		// log.Fatal("Could not get currently playing track")
	}

	// TODO implement mux
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
		fmt.Fprintf(w, "Hello there, %s \n", user.ID)

		// currentlyPlaying, err := client.PlayerCurrentlyPlaying(context.Background())
		// if err != nil {
		// 	fmt.Fprint(w, "Could not get currently playing track")
		// }

		fmt.Fprintf(w, "You are currently listening to: %s\n", currentlyPlaying.Item.Name)
	})

	done := make(chan bool)
	go forever()
	<-done // Block forever
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}

func forever() {
	for {
		// fmt.Printf("%v+\n", time.Now())
		time.Sleep(time.Second)
	}
}
