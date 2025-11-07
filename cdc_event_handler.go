package cdcelasticsearch

import (
	"gitlab.trendyol.com/order/coex/go-mongo-cdc-elasticsearch/elasticsearch/bulk"
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
