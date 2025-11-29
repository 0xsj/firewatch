-- Jobs table for background processing
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    priority INT NOT NULL DEFAULT 0,
    attempts INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    last_error TEXT,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT jobs_status_valid CHECK (status IN ('pending', 'running', 'completed', 'failed', 'retrying', 'cancelled'))
);

-- Index for dequeue query: find processable jobs ordered by priority and scheduled time
-- This is the critical index for SELECT FOR UPDATE SKIP LOCKED
CREATE INDEX idx_jobs_dequeue ON jobs (scheduled_at, priority DESC) 
    WHERE status IN ('pending', 'retrying');

-- Index for job type filtering
CREATE INDEX idx_jobs_type ON jobs (type);

-- Index for status filtering
CREATE INDEX idx_jobs_status ON jobs (status);

-- Index for cleanup: find old terminal jobs
CREATE INDEX idx_jobs_cleanup ON jobs (completed_at) 
    WHERE status IN ('completed', 'failed', 'cancelled');

-- Comments
COMMENT ON TABLE jobs IS 'Background job queue for async processing';
COMMENT ON COLUMN jobs.id IS 'Unique job identifier (UUIDv7)';
COMMENT ON COLUMN jobs.type IS 'Job type identifier (e.g., send_email, process_webhook)';
COMMENT ON COLUMN jobs.payload IS 'Job-specific data as JSON';
COMMENT ON COLUMN jobs.status IS 'Current job status';
COMMENT ON COLUMN jobs.priority IS 'Higher priority jobs are processed first';
COMMENT ON COLUMN jobs.attempts IS 'Number of processing attempts';
COMMENT ON COLUMN jobs.max_retries IS 'Maximum retry attempts before marking as failed';
COMMENT ON COLUMN jobs.last_error IS 'Error message from last failed attempt';
COMMENT ON COLUMN jobs.scheduled_at IS 'When the job should be processed (supports delayed jobs)';
COMMENT ON COLUMN jobs.started_at IS 'When processing started';
COMMENT ON COLUMN jobs.completed_at IS 'When processing finished (success or final failure)';