package db

const schemaSQL = `
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    source_path TEXT NOT NULL DEFAULT '',
    output_path TEXT NOT NULL DEFAULT '',
    progress REAL NOT NULL DEFAULT 0.0,
    estimated_eta TEXT NOT NULL DEFAULT '',
    pid INTEGER NOT NULL DEFAULT 0,
    config_id INTEGER NOT NULL DEFAULT 0,
    error_msg TEXT NOT NULL DEFAULT '',
    deleted BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS file_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    source_file TEXT NOT NULL DEFAULT '',
    output_file TEXT NOT NULL DEFAULT '',
    streams TEXT NOT NULL DEFAULT '{}',
    selected BOOLEAN NOT NULL DEFAULT 1,
    progress REAL NOT NULL DEFAULT 0.0,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transcode_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT '',
    encoder_type TEXT NOT NULL DEFAULT 'cpu',
    video_encoder TEXT NOT NULL DEFAULT 'x265',
    video_params TEXT NOT NULL DEFAULT '{}',
    audio_tracks TEXT NOT NULL DEFAULT '[]',
    subtitle_tracks TEXT NOT NULL DEFAULT '[]',
    chapters_enabled BOOLEAN NOT NULL DEFAULT 1,
    output_muxer TEXT NOT NULL DEFAULT 'mkvmerge',
    is_builtin BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS preset_templates (
    name TEXT PRIMARY KEY,
    encoder TEXT NOT NULL DEFAULT '',
    mode TEXT NOT NULL DEFAULT 'cpu',
    params TEXT NOT NULL DEFAULT '{}',
    description TEXT NOT NULL DEFAULT '',
    builtin BOOLEAN NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS preview_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL DEFAULT 0,
    source_file TEXT NOT NULL DEFAULT '',
    start_time TEXT NOT NULL DEFAULT '',
    duration INTEGER NOT NULL DEFAULT 0,
    output_file TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);
`
