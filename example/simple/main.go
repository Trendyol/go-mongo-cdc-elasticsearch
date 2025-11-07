package main

import (
	"context"

	cdcelasticsearch "github.com/Trendyol/go-mongo-cdc-elasticsearch"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch/document"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/mongodb"
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
