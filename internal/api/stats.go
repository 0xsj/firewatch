package api

import (
	"net/http"
	"sort"

	"github.com/0xsj/firewatch/internal/storage"
)

type statsResponse struct {
	TotalEvents   int64          `json:"total_events"`
	UniqueIPs     int            `json:"unique_ips"`
	ByModule      map[string]int `json:"by_module"`
	BySeverity    map[string]int `json:"by_severity"`
	TopIPs        []ipCount      `json:"top_ips"`
	TopSignatures []sigCount     `json:"top_signatures"`
}

type ipCount struct {
	IP    string `json:"ip"`
	Count int    `json:"count"`
}

type sigCount struct {
	Signature string `json:"signature"`
	Count     int    `json:"count"`
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	sinceStr := r.URL.Query().Get("since")
	if sinceStr == "" {
		sinceStr = "24h"
	}

	since, err := parseSince(sinceStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	filter := storage.EventFilter{Since: since}

	total, err := h.store.CountEvents(ctx, filter)
	if err != nil {
		h.logger.Error("api: count events", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	events, err := h.store.ListEvents(ctx, storage.EventFilter{Since: since})
	if err != nil {
		h.logger.Error("api: list events for stats", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	byModule := make(map[string]int)
	bySeverity := make(map[string]int)
	byIP := make(map[string]int)
	bySig := make(map[string]int)

	for _, e := range events {
		byModule[e.Module]++
		bySeverity[e.Severity]++
		byIP[e.SourceIP]++
		for _, sig := range e.Signatures {
			bySig[sig]++
		}
	}

	resp := statsResponse{
		TotalEvents:   total,
		UniqueIPs:     len(byIP),
		ByModule:      byModule,
		BySeverity:    bySeverity,
		TopIPs:        topIPs(byIP, 10),
		TopSignatures: topSigs(bySig, 10),
	}

	writeJSON(w, http.StatusOK, resp)
}

func topIPs(m map[string]int, limit int) []ipCount {
	result := make([]ipCount, 0, len(m))
	for k, v := range m {
		result = append(result, ipCount{IP: k, Count: v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func topSigs(m map[string]int, limit int) []sigCount {
	result := make([]sigCount, 0, len(m))
	for k, v := range m {
		result = append(result, sigCount{Signature: k, Count: v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}
