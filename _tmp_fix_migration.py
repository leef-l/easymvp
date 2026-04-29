import sqlite3
conn = sqlite3.connect(r'C:\Users\Public\project\easymvp\var\data\easymvp.db')
c = conn.cursor()
# Record 0015 migration
try:
    c.execute(
        "INSERT INTO schema_migrations(version,name,checksum,applied_at,duration_ms,status) VALUES(?,?,?,?,?,?)",
        (15, '0015_add_completion_verdict_status_fields.sql', 'manual', '2026-04-26T19:00:00', 0, 'applied')
    )
    conn.commit()
    print('Recorded 0015')
except Exception as e:
    print('Error:', e)
conn.close()
