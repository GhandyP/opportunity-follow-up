package main

import (
	"context"
	"log"
	"os"

	"github.com/GhandyP/adk-go-enterprise-cookbook/internal/adkhelper"
)

func main() {
	if err := adkhelper.Run(
		context.Background(),
		os.Args[1:],
		"case_41_opportunity_follow_up",
		"A text-only starting point for opportunity follow-up.",
		"Prepare an opportunity follow-up from account facts and last-contact text supplied by the user. Draft a next message without changing pipeline data or claiming CRM access.",
	); err != nil {
		log.Fatal(err)
	}
}
