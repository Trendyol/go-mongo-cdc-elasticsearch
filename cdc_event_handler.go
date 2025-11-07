package cdcelasticsearch

import (
	"github.com/Trendyol/go-mongo-cdc-elasticsearch/elasticsearch/bulk"
)

type CDCEventHandler struct {
	bulk *bulk.Bulk
}

func (h *CDCEventHandler) BeforeRebalanceStart() {
	h.bulk.PrepareStartRebalancing()
}

func (h *CDCEventHandler) AfterRebalanceStart() {
}

func (h *CDCEventHandler) BeforeRebalanceEnd() {
}

func (h *CDCEventHandler) AfterRebalanceEnd() {
	h.bulk.PrepareEndRebalancing()
}
