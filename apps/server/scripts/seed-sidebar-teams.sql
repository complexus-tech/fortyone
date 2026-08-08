\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE sidebar_team_seed_context ON COMMIT DROP AS
SELECT
    workspace.workspace_id,
    member.user_id
FROM workspaces AS workspace
JOIN workspace_members AS membership
  ON membership.workspace_id = workspace.workspace_id
JOIN users AS member
  ON member.user_id = membership.user_id
WHERE workspace.slug = :'workspace_slug'
  AND member.email = :'member_email';

DO $block$
BEGIN
    IF (SELECT COUNT(*) FROM sidebar_team_seed_context) <> 1 THEN
        RAISE EXCEPTION 'Sidebar team seed requires exactly one matching workspace member';
    END IF;
END
$block$;

CREATE TEMP TABLE sidebar_team_seed_data (
    name varchar(255) NOT NULL,
    code varchar(255) NOT NULL,
    color varchar(100) NOT NULL,
    order_index int4 NOT NULL
) ON COMMIT DROP;

INSERT INTO sidebar_team_seed_data (name, code, color, order_index)
VALUES
    ('Design', 'SIDEBAR-DESIGN', '#a855f7', 1000),
    ('Engineering', 'SIDEBAR-ENGINEERING', '#3b82f6', 2000),
    ('Marketing', 'SIDEBAR-MARKETING', '#f97316', 3000),
    ('Sales', 'SIDEBAR-SALES', '#eab308', 4000),
    ('Customer Success', 'SIDEBAR-CUSTOMER-SUCCESS', '#14b8a6', 5000),
    ('Operations', 'SIDEBAR-OPERATIONS', '#ec4899', 6000);

INSERT INTO teams (name, code, color, is_private, workspace_id)
SELECT
    seed.name,
    seed.code,
    seed.color,
    false,
    context.workspace_id
FROM sidebar_team_seed_context AS context
CROSS JOIN sidebar_team_seed_data AS seed
ON CONFLICT (workspace_id, code) DO NOTHING;

INSERT INTO team_members (team_id, user_id)
SELECT
    team.team_id,
    context.user_id
FROM sidebar_team_seed_context AS context
JOIN teams AS team
  ON team.workspace_id = context.workspace_id
JOIN sidebar_team_seed_data AS seed
  ON seed.code = team.code
ON CONFLICT (team_id, user_id) DO NOTHING;

INSERT INTO statuses (
    name,
    category,
    order_index,
    color,
    team_id,
    workspace_id
)
SELECT
    status.name,
    status.category,
    status.order_index,
    status.color,
    team.team_id,
    context.workspace_id
FROM sidebar_team_seed_context AS context
JOIN teams AS team
  ON team.workspace_id = context.workspace_id
JOIN sidebar_team_seed_data AS team_seed
  ON team_seed.code = team.code
CROSS JOIN (
    VALUES
        ('Backlog', 'backlog', 1000, '#6b665c'),
        ('To Do', 'unstarted', 2000, '#6b665c'),
        ('In Progress', 'started', 3000, '#eab308'),
        ('Done', 'completed', 4000, '#22c55e'),
        ('Blocked', 'paused', 5000, '#6b665c'),
        ('Cancelled', 'cancelled', 6000, '#f43f5e')
) AS status(name, category, order_index, color)
WHERE NOT EXISTS (
    SELECT 1
    FROM statuses AS existing
    WHERE existing.team_id = team.team_id
      AND existing.name = status.name
);

INSERT INTO team_story_automation_settings (team_id, workspace_id)
SELECT
    team.team_id,
    context.workspace_id
FROM sidebar_team_seed_context AS context
JOIN teams AS team
  ON team.workspace_id = context.workspace_id
JOIN sidebar_team_seed_data AS seed
  ON seed.code = team.code
ON CONFLICT (team_id, workspace_id) DO NOTHING;

INSERT INTO user_team_orders (user_id, team_id, workspace_id, order_index)
SELECT
    context.user_id,
    team.team_id,
    context.workspace_id,
    seed.order_index
FROM sidebar_team_seed_context AS context
JOIN teams AS team
  ON team.workspace_id = context.workspace_id
JOIN sidebar_team_seed_data AS seed
  ON seed.code = team.code
ON CONFLICT (user_id, team_id, workspace_id)
DO UPDATE SET
    order_index = EXCLUDED.order_index,
    updated_at = NOW();

DO $block$
BEGIN
    IF (
        SELECT COUNT(*)
        FROM sidebar_team_seed_context AS context
        JOIN teams AS team
          ON team.workspace_id = context.workspace_id
        JOIN sidebar_team_seed_data AS seed
          ON seed.code = team.code
        JOIN team_members AS membership
          ON membership.team_id = team.team_id
         AND membership.user_id = context.user_id
    ) <> 6 THEN
        RAISE EXCEPTION 'Sidebar team seed did not create all six memberships';
    END IF;
END
$block$;

SELECT
    team.name,
    team.code,
    team.color
FROM workspaces AS workspace
JOIN teams AS team
  ON team.workspace_id = workspace.workspace_id
JOIN sidebar_team_seed_data AS seed
  ON seed.code = team.code
WHERE workspace.slug = :'workspace_slug'
ORDER BY seed.order_index;

COMMIT;
