package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/0xsj/firewatch/internal/storage/models"
	"github.com/0xsj/firewatch/pkg/errors"

	_ "modernc.org/sqlite" // sqlite driver
)

const op = errors.Op("storage.sqlite")

// SQLiteStore implements Store backed by a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLite opens a SQLite database at path and runs migrations.
func NewSQLite(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, errors.E(op, errors.KindInternal, errors.CodeStorageConnect, err)
	}

	// Enable WAL mode for concurrent reads.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, errors.E(op, errors.KindInternal, errors.CodeStorageConnect, err)
	}

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) migrate() error {
	const migrationOp = errors.Op("storage.sqlite.migrate")

	_, err := s.db.Exec(schema)
	if err != nil {
		return errors.E(migrationOp, errors.KindInternal, errors.CodeStorageMigrate, err)
	}
	return nil
}

const schema = `
CREATE TABLE IF NOT EXISTS events (
	id          TEXT PRIMARY KEY,
	timestamp   TEXT NOT NULL,
	request_id  TEXT NOT NULL,
	source_ip   TEXT NOT NULL,
	source_port INTEGER NOT NULL,
	module      TEXT NOT NULL,
	method      TEXT NOT NULL,
	path        TEXT NOT NULL,
	query       TEXT,
	headers     TEXT,
	body        TEXT,
	user_agent  TEXT,
	severity    TEXT NOT NULL,
	signatures  TEXT,
	fingerprint TEXT,
	geoip       TEXT,
	reverse_dns TEXT,
	attacker_id TEXT,
	campaign_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_source_ip ON events(source_ip);
CREATE INDEX IF NOT EXISTS idx_events_module ON events(module);
CREATE INDEX IF NOT EXISTS idx_events_severity ON events(severity);

CREATE TABLE IF NOT EXISTS attackers (
	id              TEXT PRIMARY KEY,
	first_seen      TEXT NOT NULL,
	last_seen       TEXT NOT NULL,
	ip              TEXT NOT NULL UNIQUE,
	hostname        TEXT,
	geoip           TEXT,
	user_agents     TEXT,
	total_events    INTEGER NOT NULL DEFAULT 0,
	modules_targeted TEXT,
	paths_probed    TEXT,
	severity        TEXT NOT NULL,
	ja3_hashes      TEXT,
	tags            TEXT
);

CREATE INDEX IF NOT EXISTS idx_attackers_ip ON attackers(ip);
CREATE INDEX IF NOT EXISTS idx_attackers_last_seen ON attackers(last_seen);

CREATE TABLE IF NOT EXISTS campaigns (
	id              TEXT PRIMARY KEY,
	name            TEXT NOT NULL,
	first_seen      TEXT NOT NULL,
	last_seen       TEXT NOT NULL,
	attacker_ips    TEXT,
	attacker_count  INTEGER NOT NULL DEFAULT 0,
	event_count     INTEGER NOT NULL DEFAULT 0,
	modules_targeted TEXT,
	pattern         TEXT,
	severity        TEXT NOT NULL,
	tags            TEXT
);

CREATE TABLE IF NOT EXISTS iocs (
	id         TEXT PRIMARY KEY,
	type       TEXT NOT NULL,
	value      TEXT NOT NULL,
	first_seen TEXT NOT NULL,
	last_seen  TEXT NOT NULL,
	source     TEXT,
	severity   TEXT NOT NULL,
	geoip      TEXT,
	hostname   TEXT,
	tags       TEXT
);

CREATE INDEX IF NOT EXISTS idx_iocs_type ON iocs(type);
CREATE INDEX IF NOT EXISTS idx_iocs_value ON iocs(value);
`

// --- Events ---

func (s *SQLiteStore) SaveEvent(ctx context.Context, event *models.Event) error {
	headers, _ := json.Marshal(event.Headers)
	sigs, _ := json.Marshal(event.Signatures)
	fp, _ := json.Marshal(event.Fingerprint)
	geoip, _ := json.Marshal(event.GeoIP)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO events (id, timestamp, request_id, source_ip, source_port,
			module, method, path, query, headers, body, user_agent,
			severity, signatures, fingerprint, geoip, reverse_dns,
			attacker_id, campaign_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.Timestamp, event.RequestID, event.SourceIP, event.SourcePort,
		event.Module, event.Method, event.Path, event.Query,
		string(headers), event.Body, event.UserAgent,
		event.Severity, string(sigs), string(fp), string(geoip), event.ReverseDNS,
		event.AttackerID, event.CampaignID,
	)
	if err != nil {
		return errors.E(errors.Op("storage.sqlite.SaveEvent"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	return nil
}

func (s *SQLiteStore) GetEvent(ctx context.Context, id string) (*models.Event, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, timestamp, request_id, source_ip, source_port,
		module, method, path, query, headers, body, user_agent,
		severity, signatures, fingerprint, geoip, reverse_dns,
		attacker_id, campaign_id FROM events WHERE id = ?`, id)
	return scanEvent(row)
}

func (s *SQLiteStore) ListEvents(ctx context.Context, f EventFilter) ([]*models.Event, error) {
	query, args := buildEventQuery("SELECT id, timestamp, request_id, source_ip, source_port, module, method, path, query, headers, body, user_agent, severity, signatures, fingerprint, geoip, reverse_dns, attacker_id, campaign_id FROM events", f)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.E(errors.Op("storage.sqlite.ListEvents"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *SQLiteStore) CountEvents(ctx context.Context, f EventFilter) (int64, error) {
	query, args := buildEventQuery("SELECT COUNT(*) FROM events", f)

	var count int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, errors.E(errors.Op("storage.sqlite.CountEvents"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	return count, nil
}

// --- Attackers ---

func (s *SQLiteStore) SaveAttacker(ctx context.Context, a *models.Attacker) error {
	geoip, _ := json.Marshal(a.GeoIP)
	userAgents, _ := json.Marshal(a.UserAgents)
	modules, _ := json.Marshal(a.ModulesTargeted)
	paths, _ := json.Marshal(a.PathsProbed)
	ja3, _ := json.Marshal(a.JA3Hashes)
	tags, _ := json.Marshal(a.Tags)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO attackers (id, first_seen, last_seen, ip, hostname, geoip,
			user_agents, total_events, modules_targeted, paths_probed,
			severity, ja3_hashes, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ip) DO UPDATE SET
			last_seen = excluded.last_seen,
			hostname = excluded.hostname,
			geoip = excluded.geoip,
			user_agents = excluded.user_agents,
			total_events = excluded.total_events,
			modules_targeted = excluded.modules_targeted,
			paths_probed = excluded.paths_probed,
			severity = excluded.severity,
			ja3_hashes = excluded.ja3_hashes,
			tags = excluded.tags`,
		a.ID, a.FirstSeen, a.LastSeen, a.IP, a.Hostname, string(geoip),
		string(userAgents), a.TotalEvents, string(modules), string(paths),
		a.Severity, string(ja3), string(tags),
	)
	if err != nil {
		return errors.E(errors.Op("storage.sqlite.SaveAttacker"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	return nil
}

func (s *SQLiteStore) GetAttacker(ctx context.Context, id string) (*models.Attacker, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, first_seen, last_seen, ip, hostname, geoip,
		user_agents, total_events, modules_targeted, paths_probed,
		severity, ja3_hashes, tags FROM attackers WHERE id = ?`, id)
	return scanAttacker(row)
}

func (s *SQLiteStore) GetAttackerByIP(ctx context.Context, ip string) (*models.Attacker, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, first_seen, last_seen, ip, hostname, geoip,
		user_agents, total_events, modules_targeted, paths_probed,
		severity, ja3_hashes, tags FROM attackers WHERE ip = ?`, ip)
	return scanAttacker(row)
}

func (s *SQLiteStore) ListAttackers(ctx context.Context, f AttackerFilter) ([]*models.Attacker, error) {
	var where []string
	var args []any

	if !f.Since.IsZero() {
		where = append(where, "last_seen >= ?")
		args = append(args, f.Since.Format(time.RFC3339))
	}
	if !f.Until.IsZero() {
		where = append(where, "first_seen <= ?")
		args = append(args, f.Until.Format(time.RFC3339))
	}
	if f.Severity != "" {
		where = append(where, "severity = ?")
		args = append(args, f.Severity)
	}

	query := "SELECT id, first_seen, last_seen, ip, hostname, geoip, user_agents, total_events, modules_targeted, paths_probed, severity, ja3_hashes, tags FROM attackers"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY last_seen DESC"
	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
	}
	if f.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", f.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.E(errors.Op("storage.sqlite.ListAttackers"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	defer rows.Close()

	var attackers []*models.Attacker
	for rows.Next() {
		a, err := scanAttacker(rows)
		if err != nil {
			return nil, err
		}
		attackers = append(attackers, a)
	}
	return attackers, rows.Err()
}

// --- Campaigns ---

func (s *SQLiteStore) SaveCampaign(ctx context.Context, c *models.Campaign) error {
	ips, _ := json.Marshal(c.AttackerIPs)
	modules, _ := json.Marshal(c.ModulesTargeted)
	tags, _ := json.Marshal(c.Tags)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO campaigns (id, name, first_seen, last_seen, attacker_ips,
			attacker_count, event_count, modules_targeted, pattern, severity, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			last_seen = excluded.last_seen,
			attacker_ips = excluded.attacker_ips,
			attacker_count = excluded.attacker_count,
			event_count = excluded.event_count,
			modules_targeted = excluded.modules_targeted,
			pattern = excluded.pattern,
			severity = excluded.severity,
			tags = excluded.tags`,
		c.ID, c.Name, c.FirstSeen, c.LastSeen, string(ips),
		c.AttackerCount, c.EventCount, string(modules), c.Pattern, c.Severity, string(tags),
	)
	if err != nil {
		return errors.E(errors.Op("storage.sqlite.SaveCampaign"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	return nil
}

func (s *SQLiteStore) GetCampaign(ctx context.Context, id string) (*models.Campaign, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, first_seen, last_seen, attacker_ips,
		attacker_count, event_count, modules_targeted, pattern, severity, tags
		FROM campaigns WHERE id = ?`, id)
	return scanCampaign(row)
}

func (s *SQLiteStore) ListCampaigns(ctx context.Context, f CampaignFilter) ([]*models.Campaign, error) {
	var where []string
	var args []any

	if !f.Since.IsZero() {
		where = append(where, "last_seen >= ?")
		args = append(args, f.Since.Format(time.RFC3339))
	}
	if !f.Until.IsZero() {
		where = append(where, "first_seen <= ?")
		args = append(args, f.Until.Format(time.RFC3339))
	}

	query := "SELECT id, name, first_seen, last_seen, attacker_ips, attacker_count, event_count, modules_targeted, pattern, severity, tags FROM campaigns"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY last_seen DESC"
	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
	}
	if f.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", f.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.E(errors.Op("storage.sqlite.ListCampaigns"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	defer rows.Close()

	var campaigns []*models.Campaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

// --- IOCs ---

func (s *SQLiteStore) SaveIOC(ctx context.Context, ioc *models.IOC) error {
	geoip, _ := json.Marshal(ioc.GeoIP)
	tags, _ := json.Marshal(ioc.Tags)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO iocs (id, type, value, first_seen, last_seen,
			source, severity, geoip, hostname, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			last_seen = excluded.last_seen,
			severity = excluded.severity,
			geoip = excluded.geoip,
			hostname = excluded.hostname,
			tags = excluded.tags`,
		ioc.ID, ioc.Type, ioc.Value, ioc.FirstSeen, ioc.LastSeen,
		ioc.Source, ioc.Severity, string(geoip), ioc.Hostname, string(tags),
	)
	if err != nil {
		return errors.E(errors.Op("storage.sqlite.SaveIOC"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	return nil
}

func (s *SQLiteStore) ListIOCs(ctx context.Context, f IOCFilter) ([]*models.IOC, error) {
	var where []string
	var args []any

	if !f.Since.IsZero() {
		where = append(where, "last_seen >= ?")
		args = append(args, f.Since.Format(time.RFC3339))
	}
	if !f.Until.IsZero() {
		where = append(where, "first_seen <= ?")
		args = append(args, f.Until.Format(time.RFC3339))
	}
	if f.Type != "" {
		where = append(where, "type = ?")
		args = append(args, f.Type)
	}
	if f.Severity != "" {
		where = append(where, "severity = ?")
		args = append(args, f.Severity)
	}

	query := "SELECT id, type, value, first_seen, last_seen, source, severity, geoip, hostname, tags FROM iocs"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY last_seen DESC"
	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
	}
	if f.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", f.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.E(errors.Op("storage.sqlite.ListIOCs"), errors.KindInternal, errors.CodeStorageQuery, err)
	}
	defer rows.Close()

	var iocs []*models.IOC
	for rows.Next() {
		ioc, err := scanIOC(rows)
		if err != nil {
			return nil, err
		}
		iocs = append(iocs, ioc)
	}
	return iocs, rows.Err()
}

// --- Lifecycle ---

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// --- Scanners ---

// scanner abstracts *sql.Row and *sql.Rows for reuse.
type scanner interface {
	Scan(dest ...any) error
}

func scanEvent(s scanner) (*models.Event, error) {
	var e models.Event
	var headers, sigs, fp, geoip sql.NullString

	err := s.Scan(
		&e.ID, &e.Timestamp, &e.RequestID, &e.SourceIP, &e.SourcePort,
		&e.Module, &e.Method, &e.Path, &e.Query,
		&headers, &e.Body, &e.UserAgent,
		&e.Severity, &sigs, &fp, &geoip, &e.ReverseDNS,
		&e.AttackerID, &e.CampaignID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(errors.Op("storage.sqlite.scanEvent"), errors.KindNotFound, "event not found")
		}
		return nil, errors.E(errors.Op("storage.sqlite.scanEvent"), errors.KindInternal, errors.CodeStorageQuery, err)
	}

	if headers.Valid {
		_ = json.Unmarshal([]byte(headers.String), &e.Headers)
	}
	if sigs.Valid {
		_ = json.Unmarshal([]byte(sigs.String), &e.Signatures)
	}
	if fp.Valid {
		_ = json.Unmarshal([]byte(fp.String), &e.Fingerprint)
	}
	if geoip.Valid {
		_ = json.Unmarshal([]byte(geoip.String), &e.GeoIP)
	}

	return &e, nil
}

func scanAttacker(s scanner) (*models.Attacker, error) {
	var a models.Attacker
	var geoip, userAgents, modules, paths, ja3, tags sql.NullString

	err := s.Scan(
		&a.ID, &a.FirstSeen, &a.LastSeen, &a.IP, &a.Hostname, &geoip,
		&userAgents, &a.TotalEvents, &modules, &paths,
		&a.Severity, &ja3, &tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(errors.Op("storage.sqlite.scanAttacker"), errors.KindNotFound, "attacker not found")
		}
		return nil, errors.E(errors.Op("storage.sqlite.scanAttacker"), errors.KindInternal, errors.CodeStorageQuery, err)
	}

	if geoip.Valid {
		_ = json.Unmarshal([]byte(geoip.String), &a.GeoIP)
	}
	if userAgents.Valid {
		_ = json.Unmarshal([]byte(userAgents.String), &a.UserAgents)
	}
	if modules.Valid {
		_ = json.Unmarshal([]byte(modules.String), &a.ModulesTargeted)
	}
	if paths.Valid {
		_ = json.Unmarshal([]byte(paths.String), &a.PathsProbed)
	}
	if ja3.Valid {
		_ = json.Unmarshal([]byte(ja3.String), &a.JA3Hashes)
	}
	if tags.Valid {
		_ = json.Unmarshal([]byte(tags.String), &a.Tags)
	}

	return &a, nil
}

func scanCampaign(s scanner) (*models.Campaign, error) {
	var c models.Campaign
	var ips, modules, tags sql.NullString

	err := s.Scan(
		&c.ID, &c.Name, &c.FirstSeen, &c.LastSeen, &ips,
		&c.AttackerCount, &c.EventCount, &modules, &c.Pattern, &c.Severity, &tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(errors.Op("storage.sqlite.scanCampaign"), errors.KindNotFound, "campaign not found")
		}
		return nil, errors.E(errors.Op("storage.sqlite.scanCampaign"), errors.KindInternal, errors.CodeStorageQuery, err)
	}

	if ips.Valid {
		_ = json.Unmarshal([]byte(ips.String), &c.AttackerIPs)
	}
	if modules.Valid {
		_ = json.Unmarshal([]byte(modules.String), &c.ModulesTargeted)
	}
	if tags.Valid {
		_ = json.Unmarshal([]byte(tags.String), &c.Tags)
	}

	return &c, nil
}

func scanIOC(s scanner) (*models.IOC, error) {
	var ioc models.IOC
	var geoip, tags sql.NullString
	var iocType string

	err := s.Scan(
		&ioc.ID, &iocType, &ioc.Value, &ioc.FirstSeen, &ioc.LastSeen,
		&ioc.Source, &ioc.Severity, &geoip, &ioc.Hostname, &tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.E(errors.Op("storage.sqlite.scanIOC"), errors.KindNotFound, "IOC not found")
		}
		return nil, errors.E(errors.Op("storage.sqlite.scanIOC"), errors.KindInternal, errors.CodeStorageQuery, err)
	}

	ioc.Type = models.IOCType(iocType)
	if geoip.Valid {
		_ = json.Unmarshal([]byte(geoip.String), &ioc.GeoIP)
	}
	if tags.Valid {
		_ = json.Unmarshal([]byte(tags.String), &ioc.Tags)
	}

	return &ioc, nil
}

// --- Query Builder ---

func buildEventQuery(base string, f EventFilter) (string, []any) {
	var where []string
	var args []any

	if !f.Since.IsZero() {
		where = append(where, "timestamp >= ?")
		args = append(args, f.Since.Format(time.RFC3339))
	}
	if !f.Until.IsZero() {
		where = append(where, "timestamp <= ?")
		args = append(args, f.Until.Format(time.RFC3339))
	}
	if f.SourceIP != "" {
		where = append(where, "source_ip = ?")
		args = append(args, f.SourceIP)
	}
	if f.Module != "" {
		where = append(where, "module = ?")
		args = append(args, f.Module)
	}
	if f.Severity != "" {
		where = append(where, "severity = ?")
		args = append(args, f.Severity)
	}

	query := base
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY timestamp DESC"
	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
	}
	if f.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", f.Offset)
	}

	return query, args
}
