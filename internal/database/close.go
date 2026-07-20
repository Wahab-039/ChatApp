package database

// Close releases all connections held by the PostgreSQL pool.
func (p *Postgres) Close() {
	if p == nil || p.Pool == nil {
		return
	}

	p.Pool.Close()
}
