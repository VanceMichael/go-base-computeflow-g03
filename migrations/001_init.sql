CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL);

CREATE TABLE IF NOT EXISTS ports (
  id TEXT PRIMARY KEY, code TEXT NOT NULL UNIQUE, name TEXT NOT NULL,
  timezone TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  email TEXT NOT NULL, display_name TEXT NOT NULL, role TEXT NOT NULL,
  active INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL,
  UNIQUE(port_id,email)
);
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id),
  token_hash TEXT NOT NULL UNIQUE, expires_at TEXT NOT NULL,
  revoked_at TEXT, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS stress_runs (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  name TEXT NOT NULL, state TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1,
  starts_at TEXT NOT NULL, ends_at TEXT, created_by TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS passenger_waves (
  id TEXT PRIMARY KEY, run_id TEXT NOT NULL REFERENCES stress_runs(id),
  sequence_no INTEGER NOT NULL, state TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1,
  planned_at TEXT NOT NULL, released_at TEXT, UNIQUE(run_id,sequence_no)
);
CREATE TABLE IF NOT EXISTS passengers (
  id TEXT PRIMARY KEY, wave_id TEXT NOT NULL REFERENCES passenger_waves(id),
  document_key TEXT NOT NULL, state TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL, UNIQUE(wave_id,document_key)
);
CREATE TABLE IF NOT EXISTS gates (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  gate_no INTEGER NOT NULL, mode TEXT NOT NULL, active INTEGER NOT NULL DEFAULT 1,
  UNIQUE(port_id,gate_no)
);
CREATE TABLE IF NOT EXISTS gate_scans (
  id TEXT PRIMARY KEY, passenger_id TEXT NOT NULL REFERENCES passengers(id),
  gate_id TEXT NOT NULL REFERENCES gates(id), stage INTEGER NOT NULL,
  state TEXT NOT NULL, lease_owner TEXT, lease_token TEXT, lease_until TEXT,
  version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL,
  UNIQUE(passenger_id,stage)
);
CREATE TABLE IF NOT EXISTS vehicle_batches (
  id TEXT PRIMARY KEY, run_id TEXT NOT NULL REFERENCES stress_runs(id),
  manifest_key TEXT NOT NULL, state TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL, UNIQUE(run_id,manifest_key)
);
CREATE TABLE IF NOT EXISTS vehicles (
  id TEXT PRIMARY KEY, batch_id TEXT NOT NULL REFERENCES vehicle_batches(id),
  plate_key TEXT NOT NULL, state TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL, UNIQUE(batch_id,plate_key)
);
CREATE TABLE IF NOT EXISTS lanes (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  lane_no INTEGER NOT NULL, state TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1,
  UNIQUE(port_id,lane_no)
);
CREATE TABLE IF NOT EXISTS lane_assignments (
  id TEXT PRIMARY KEY, lane_id TEXT NOT NULL REFERENCES lanes(id),
  vehicle_id TEXT NOT NULL REFERENCES vehicles(id), owner TEXT NOT NULL,
  state TEXT NOT NULL, lease_until TEXT, version INTEGER NOT NULL DEFAULT 1,
  UNIQUE(lane_id), UNIQUE(vehicle_id)
);
CREATE TABLE IF NOT EXISTS risk_assessments (
  id TEXT PRIMARY KEY, subject_type TEXT NOT NULL, subject_id TEXT NOT NULL,
  state TEXT NOT NULL, reason TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS incidents (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  run_id TEXT REFERENCES stress_runs(id), subject_type TEXT NOT NULL,
  subject_id TEXT NOT NULL, state TEXT NOT NULL, severity TEXT NOT NULL,
  description TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS responders (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  name TEXT NOT NULL, state TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS dispatch_assignments (
  id TEXT PRIMARY KEY, incident_id TEXT NOT NULL REFERENCES incidents(id),
  responder_id TEXT NOT NULL REFERENCES responders(id), owner TEXT NOT NULL,
  state TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, UNIQUE(incident_id), UNIQUE(responder_id)
);
CREATE TABLE IF NOT EXISTS outbox_messages (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  event_key TEXT NOT NULL, payload TEXT NOT NULL, state TEXT NOT NULL,
  owner TEXT, lease_until TEXT, attempts INTEGER NOT NULL DEFAULT 0,
  idempotency_key TEXT NOT NULL, last_error TEXT, created_at TEXT NOT NULL,
  delivered_at TEXT, UNIQUE(port_id,idempotency_key)
);
CREATE TABLE IF NOT EXISTS capacity_snapshots (
  id TEXT PRIMARY KEY, run_id TEXT NOT NULL REFERENCES stress_runs(id),
  window_start TEXT NOT NULL, window_end TEXT NOT NULL, passengers INTEGER NOT NULL,
  cleared INTEGER NOT NULL, held INTEGER NOT NULL, vehicles INTEGER NOT NULL,
  state TEXT NOT NULL, created_at TEXT NOT NULL, UNIQUE(run_id,window_start,window_end)
);
CREATE TABLE IF NOT EXISTS audit_events (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  actor_id TEXT, action TEXT NOT NULL, subject_type TEXT NOT NULL, subject_id TEXT NOT NULL,
  outcome TEXT NOT NULL, request_id TEXT NOT NULL, details TEXT NOT NULL, created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS idempotency_records (
  id TEXT PRIMARY KEY, port_id TEXT NOT NULL REFERENCES ports(id),
  operation TEXT NOT NULL, request_key TEXT NOT NULL, response_code INTEGER NOT NULL,
  response_body TEXT NOT NULL, created_at TEXT NOT NULL, UNIQUE(port_id,operation,request_key)
);

CREATE INDEX IF NOT EXISTS idx_passengers_wave_state ON passengers(wave_id,state);
CREATE INDEX IF NOT EXISTS idx_incidents_port_state ON incidents(port_id,state);
CREATE INDEX IF NOT EXISTS idx_outbox_state_lease ON outbox_messages(state,lease_until);
CREATE INDEX IF NOT EXISTS idx_audit_port_time ON audit_events(port_id,created_at);
