package main

import (
	"context"

	cdcelasticsearch "gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch"
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/elasticsearch/document"
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/mongodb"
)

func mapper(event mongodb.Event) []document.ESActionDocument {
	if event.IsMutated {
		e := document.NewIndexAction(event.Key, event.Value, nil)
		return []document.ESActionDocument{e}
	}
	e := document.NewDeleteAction(event.Key, nil)
	return []document.ESActionDocument{e}
}

func main() {
	connector, err := cdcelasticsearch.NewConnectorBuilder("config.yml").
		SetMapper(mapper).
		Build()
	if err != nil {
		panic(err)
	}

	defer connector.Close()
	connector.Start(context.Background())
}
