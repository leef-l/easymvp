import sqlite3
conn = sqlite3.connect(r'C:\Users\Public\project\easymvp\var\data\easymvp.db')
c = conn.cursor()
c.execute("SELECT name FROM schema_migrations WHERE name = '0015_add_completion_verdict_status_fields.sql'")
row = c.fetchone()
print('0015 in schema_migrations:', row)
c.execute('PRAGMA table_info(completion_verdicts)')
cols = [r[1] for r in c.fetchall()]
print('Columns:', cols)
conn.close()
