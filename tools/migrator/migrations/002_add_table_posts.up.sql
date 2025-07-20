CREATE TABLE IF NOT EXISTS posts (
		id UUID PRIMARY KEY,
        owner_id UUID NOT NULL,
        header TEXT NOT NULL,
        text TEXT NOT NULL,
        price INTEGER NOT NULL,
        created_at TIMESTAMP,
        FOREIGN KEY(owner_id) REFERENCES users(id)
        );