package main

import (
	"context"

	cdcelasticsearch "gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch"
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
