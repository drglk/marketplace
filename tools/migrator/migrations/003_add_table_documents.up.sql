CREATE TABLE IF NOT EXISTS documents (
		id UUID PRIMARY KEY,
        post_id UUID NOT NULL,
        name TEXT NOT NULL,
        mime TEXT NOT NULL,
        path TEXT,
        FOREIGN KEY(post_id) REFERENCES posts(id)
        );