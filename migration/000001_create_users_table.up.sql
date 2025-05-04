CREATE TABLE IF NOT EXISTS users (
                                     id VARCHAR(36) PRIMARY KEY,
                                     email VARCHAR(255) UNIQUE NOT NULL,
                                     password VARCHAR(255) NOT NULL,
                                     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                     updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_codes (
                                          user_id VARCHAR(36) REFERENCES users(id),
                                          code VARCHAR(4) NOT NULL,
                                          expires_at TIMESTAMP NOT NULL,
                                          PRIMARY KEY (user_id, code)
);