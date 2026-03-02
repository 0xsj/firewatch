package api

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/storage"
	"github.com/0xsj/firewatch/internal/storage/models"
)

func (h *Handler) handleTokens(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	since, err := parseSince(q.Get("since"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := storage.HoneyTokenFilter{
		Kind:     q.Get("kind"),
		SourceIP: q.Get("ip"),
		Module:   q.Get("module"),
		Since:    since,
		Limit:    parseIntParam(r, "limit", 50),
		Offset:   parseIntParam(r, "offset", 0),
	}

	tokens, err := h.store.ListHoneyTokens(r.Context(), filter)
	if err != nil {
		h.logger.Error("api: list tokens", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if tokens == nil {
		tokens = make([]*models.HoneyToken, 0)
	}

	writeJSON(w, http.StatusOK, tokens)
}
