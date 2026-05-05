CREATE TABLE IF NOT EXISTS payments (
                                        id TEXT PRIMARY KEY,
                                        order_id TEXT NOT NULL,
                                        transaction_id TEXT NOT NULL,
                                        customer_email TEXT NOT NULL,
                                        amount BIGINT NOT NULL,
                                        status TEXT NOT NULL
);