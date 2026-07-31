package db

import "database/sql"

func GetSetting(db *sql.DB, key string) (string, error) {
	var val string
	err := db.QueryRow("SELECT value FROM system_settings WHERE key=?", key).Scan(&val)
	return val, err
}

func SetSetting(db *sql.DB, key, value string) error {
	_, err := db.Exec("INSERT INTO system_settings (key, value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", key, value)
	return err
}

func ListSettings(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT key, value FROM system_settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	settings := map[string]string{}
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		settings[k] = v
	}
	return settings, nil
}
