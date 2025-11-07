package main

import (
	"context"

	cdcelasticsearch "github.com/Trendyol/go-mongo-cdc-elasticsearch"
)

func main() {
	connector, err := cdcelasticsearch.NewConnectorBuilder("config.yml").
		Build()
	if err != nil {
		panic(err)
	}

	defer connector.Close()
	connector.Start(context.Background())
}
