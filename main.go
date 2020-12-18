package main

import (
	"fmt"
	"time"

	plank "github.com/armory/plank/v3"
)

type KayentaThing struct {
	ID   string `json:"id,omitempty" yaml:"id,omitempty" hcl:"id,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty" hcl:"type,omitempty"`
}

func main() {
	fmt.Printf("somestuff")
	c := plank.New(plank.WithMaxRetries(1), plank.WithRetryIncrement(time.Second))
	var things []KayentaThing
	if err := c.Get("https://a5c6ffa5-3028-450c-a1f9-99de7e439480.mock.pstmn.io/users/123", &things); err != nil {
		fmt.Println("could not get pipelines")
		fmt.Println(err)
	} else {
		fmt.Println("somethings")
	}
}
