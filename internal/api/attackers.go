package api

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

func (h *Handler) handleAttackers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	since, err := parseSince(q.Get("since"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := storage.AttackerFilter{
		Severity: q.Get("severity"),
		Tag:      q.Get("tag"),
		Since:    since,
		Limit:    parseIntParam(r, "limit", 50),
		Offset:   parseIntParam(r, "offset", 0),
	}

	attackers, err := h.store.ListAttackers(r.Context(), filter)
	if err != nil {
		h.logger.Error("api: list attackers", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if attackers == nil {
		attackers = make([]*models.Attacker, 0)
	}

	writeJSON(w, http.StatusOK, attackers)
}
