DROP INDEX IF EXISTS issue_status_events_issue_id_idx;
DROP TABLE IF EXISTS issue_status_events;
ALTER TABLE issues DROP COLUMN changed_reason;
