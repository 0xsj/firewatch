package api

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	since, err := parseSince(q.Get("since"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := storage.EventFilter{
		Module:   q.Get("module"),
		SourceIP: q.Get("ip"),
		Severity: q.Get("severity"),
		Since:    since,
		Limit:    parseIntParam(r, "limit", 50),
		Offset:   parseIntParam(r, "offset", 0),
	}

	events, err := h.store.ListEvents(r.Context(), filter)
	if err != nil {
		h.logger.Error("api: list events", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if events == nil {
		events = make([]*models.Event, 0)
	}

	writeJSON(w, http.StatusOK, events)
}
