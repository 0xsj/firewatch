package exposure

import (
	"net/http"

	"github.com/0xsj/firewatch/internal/handlers"
	"github.com/0xsj/firewatch/pkg/httputil"
)

const (
	sigGitProbe  = "exposure-git-001"
	sigGitConfig = "exposure-git-002"
	sigGitHEAD   = "exposure-git-003"
)

func (e *Exposure) handleGit(w http.ResponseWriter, r *http.Request) {
	sigs := []string{sigGitProbe}
	severity := "high"

	switch r.URL.Path {
	case "/.git/config":
		sigs = append(sigs, sigGitConfig)
		e.logger.Info("git config probe", "ip", httputil.ClientIP(r))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[core]\n\trepositoryformatversion = 0\n\tfilemode = true\n\tbare = false\n[remote \"origin\"]\n\turl = git@github.com:example/app.git\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n"))

	case "/.git/HEAD":
		sigs = append(sigs, sigGitHEAD)
		e.logger.Info("git HEAD probe", "ip", httputil.ClientIP(r))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ref: refs/heads/main\n"))

	default:
		e.logger.Info("git directory probe",
			"path", r.URL.Path,
			"ip", httputil.ClientIP(r),
		)
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	}

	handlers.RecordEvent(e.store, e.logger, r, moduleName, severity, sigs)
}
