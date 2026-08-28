CREATE TYPE sqlc_contract_status AS ENUM ('active', 'disabled');

CREATE TABLE sqlc_type_contracts (
    id uuid PRIMARY KEY,
    nullable_id uuid,
    occurred_at timestamptz NOT NULL,
    nullable_occurred_at timestamptz,
    local_at timestamp NOT NULL,
    nullable_local_at timestamp,
    due_date date NOT NULL,
    nullable_due_date date,
    status sqlc_contract_status NOT NULL,
    nullable_status sqlc_contract_status,
    amount numeric NOT NULL,
    payload jsonb NOT NULL,
    related_ids uuid[] NOT NULL
);
