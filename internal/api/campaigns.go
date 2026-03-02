package api

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

func (h *Handler) handleCampaigns(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	since, err := parseSince(q.Get("since"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := storage.CampaignFilter{
		Active: q.Get("active") == "true",
		Since:  since,
		Limit:  parseIntParam(r, "limit", 50),
		Offset: parseIntParam(r, "offset", 0),
	}

	campaigns, err := h.store.ListCampaigns(r.Context(), filter)
	if err != nil {
		h.logger.Error("api: list campaigns", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if campaigns == nil {
		campaigns = make([]*models.Campaign, 0)
	}

	writeJSON(w, http.StatusOK, campaigns)
}
