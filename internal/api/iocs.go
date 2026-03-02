package api

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

func (h *Handler) handleIOCs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	since, err := parseSince(q.Get("since"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := storage.IOCFilter{
		Type:     models.IOCType(q.Get("type")),
		Severity: q.Get("severity"),
		Since:    since,
		Limit:    parseIntParam(r, "limit", 100),
		Offset:   parseIntParam(r, "offset", 0),
	}

	iocs, err := h.store.ListIOCs(r.Context(), filter)
	if err != nil {
		h.logger.Error("api: list iocs", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if iocs == nil {
		iocs = make([]*models.IOC, 0)
	}

	writeJSON(w, http.StatusOK, iocs)
}
