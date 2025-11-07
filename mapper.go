package cdcelasticsearch

import (
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/elasticsearch/document"
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/mongodb"
)

type Mapper func(event mongodb.Event) []document.ESActionDocument

func DefaultMapper(event mongodb.Event) []document.ESActionDocument {
	if event.IsMutated {
		return []document.ESActionDocument{document.NewIndexAction(event.Key, event.Value, nil)}
	}
	return []document.ESActionDocument{document.NewDeleteAction(event.Key, nil)}
}
