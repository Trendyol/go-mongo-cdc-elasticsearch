package cdcelasticsearch

import (
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch/document"
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/mongodb"
)

type Mapper func(event mongodb.Event) []document.ESActionDocument

func DefaultMapper(event mongodb.Event) []document.ESActionDocument {
	if event.IsMutated {
		return []document.ESActionDocument{document.NewIndexAction(event.Key, event.Value, nil)}
	}
	return []document.ESActionDocument{document.NewDeleteAction(event.Key, nil)}
}
