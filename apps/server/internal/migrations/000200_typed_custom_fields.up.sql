CREATE TABLE custom_fields (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    team_id uuid NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
    name varchar(100) NOT NULL CHECK (length(btrim(name)) > 0),
    field_type text NOT NULL CHECK (field_type IN ('text','number','money','date','select','person')),
    currency varchar(3),
    show_on_create boolean NOT NULL DEFAULT false,
    archived_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id, workspace_id, team_id, field_type),
    CHECK ((field_type = 'money' AND currency ~ '^[A-Z]{3}$') OR (field_type <> 'money' AND currency IS NULL))
);
CREATE UNIQUE INDEX custom_fields_active_name ON custom_fields (team_id, lower(name)) WHERE archived_at IS NULL;
CREATE INDEX custom_fields_team ON custom_fields (workspace_id, team_id, created_at, id);

CREATE TABLE custom_field_options (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id uuid NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL CHECK (length(btrim(name)) > 0),
    position integer NOT NULL CHECK (position >= 0),
    archived_at timestamptz,
    UNIQUE (id, field_id)
);
CREATE UNIQUE INDEX custom_field_options_active_name ON custom_field_options (field_id, lower(name)) WHERE archived_at IS NULL;

CREATE TABLE story_custom_field_values (
    story_id uuid NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    field_id uuid NOT NULL,
    workspace_id uuid NOT NULL,
    team_id uuid NOT NULL,
    field_type text NOT NULL,
    text_value text,
    numeric_value numeric,
    date_value date,
    option_id uuid,
    person_id uuid REFERENCES users(user_id) ON DELETE SET NULL,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (story_id, field_id),
    FOREIGN KEY (field_id, workspace_id, team_id, field_type) REFERENCES custom_fields(id, workspace_id, team_id, field_type) ON DELETE CASCADE,
    FOREIGN KEY (option_id, field_id) REFERENCES custom_field_options(id, field_id),
    CHECK (num_nonnulls(text_value, numeric_value, date_value, option_id, person_id) <= 1),
    CHECK (text_value IS NULL OR field_type = 'text'),
    CHECK (numeric_value IS NULL OR field_type IN ('number','money')),
    CHECK (date_value IS NULL OR field_type = 'date'),
    CHECK (option_id IS NULL OR field_type = 'select'),
    CHECK (person_id IS NULL OR field_type = 'person')
);
CREATE INDEX story_custom_field_values_report ON story_custom_field_values (workspace_id, team_id, field_id, story_id);

ALTER TABLE stories ADD COLUMN custom_fields_version bigint NOT NULL DEFAULT 0;
CREATE TABLE story_custom_field_audit (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id uuid NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    field_id uuid NOT NULL REFERENCES custom_fields(id) ON DELETE CASCADE,
    workspace_id uuid NOT NULL REFERENCES workspaces(workspace_id) ON DELETE CASCADE,
    actor_id uuid REFERENCES users(user_id) ON DELETE SET NULL,
    old_value text,
    new_value text,
    version bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX story_custom_field_audit_story ON story_custom_field_audit (workspace_id, story_id, created_at, id);
