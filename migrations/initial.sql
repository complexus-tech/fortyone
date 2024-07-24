-- Roles
INSERT INTO roles (name, permissions)
VALUES (
    'Admin',
    '{"can_manage_users": true, "can_manage_projects": true, "can_manage_teams": true}'
  ),
  (
    'User',
    '{"can_manage_users": false, "can_manage_projects": true, "can_manage_teams": true}'
  );
-- Users
INSERT INTO users (
    username,
    email,
    password_hash,
    full_name,
    role_id,
    avatar_url,
    is_active,
    last_login_at
  )
VALUES (
    'admin_user',
    'admin@example.com',
    'hash_password',
    'Admin User',
    (
      SELECT role_id
      FROM roles
      WHERE name = 'Admin'
    ),
    'https://example.com/avatar1.png',
    true,
    CURRENT_TIMESTAMP
  ),
  (
    'regular_user',
    'user@example.com',
    'hash_password',
    'Regular User',
    (
      SELECT role_id
      FROM roles
      WHERE name = 'User'
    ),
    'https://example.com/avatar2.png',
    true,
    CURRENT_TIMESTAMP
  );
-- Workspaces
INSERT INTO workspaces (name, description)
VALUES ('Workspace 1', 'This is the first workspace'),
  ('Workspace 2', 'This is the second workspace');
-- Teams
INSERT INTO teams (
    name,
    description,
    workspace_id,
    code,
    color,
    icon
  )
VALUES (
    'Team Alpha',
    'Alpha team description',
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    ),
    'ALPHA',
    '#FF5733',
    'alpha_icon.png'
  ),
  (
    'Team Beta',
    'Beta team description',
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    ),
    'BETA',
    '#33FF57',
    'beta_icon.png'
  );
-- Objectives
INSERT INTO objectives (
    name,
    description,
    lead_user_id,
    team_id,
    workspace_id,
    start_date,
    end_date,
    status,
    is_private,
    visibility
  )
VALUES (
    'Objective 1',
    'Objective 1 description',
    (
      SELECT user_id
      FROM users
      WHERE username = 'admin_user'
    ),
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    ),
    '2024-01-01',
    '2024-06-30',
    'started',
    false,
    'public'
  ),
  (
    'Objective 2',
    'Objective 2 description',
    (
      SELECT user_id
      FROM users
      WHERE username = 'regular_user'
    ),
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    ),
    '2024-02-01',
    '2024-07-31',
    'unstarted',
    true,
    'private'
  );
-- Statuses
INSERT INTO statuses (
    name,
    color,
    category,
    order_index,
    team_id,
    workspace_id
  )
VALUES (
    'Backlog',
    '#FF5733',
    'backlog',
    1,
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    )
  ),
  (
    'Unstarted',
    '#33FF57',
    'unstarted',
    2,
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    )
  ),
  (
    'Started',
    '#5733FF',
    'started',
    3,
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    )
  ),
  (
    'Paused',
    '#F1C40F',
    'paused',
    4,
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    )
  ),
  (
    'Completed',
    '#1ABC9C',
    'completed',
    5,
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    )
  ),
  (
    'Cancelled',
    '#E74C3C',
    'cancelled',
    6,
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    )
  );
-- Sprints
INSERT INTO sprints (
    objective_id,
    team_id,
    workspace_id,
    name,
    start_date,
    end_date,
    goal,
    status,
    backlog_status,
    completed_story_count
  )
VALUES (
    (
      SELECT objective_id
      FROM objectives
      WHERE name = 'Objective 1'
    ),
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    ),
    'Sprint 1',
    '2024-01-01',
    '2024-01-14',
    'Sprint 1 goal',
    'started',
    'backlog',
    3
  ),
  (
    (
      SELECT objective_id
      FROM objectives
      WHERE name = 'Objective 2'
    ),
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    ),
    'Sprint 2',
    '2024-02-01',
    '2024-02-14',
    'Sprint 2 goal',
    'unstarted',
    'unstarted',
    0
  );
-- Attachments
INSERT INTO attachments (
    filename,
    path,
    size,
    mime_type,
    uploaded_by,
    team_id,
    workspace_id
  )
VALUES (
    'file1.txt',
    '/path/to/file1.txt',
    12345,
    'text/plain',
    (
      SELECT user_id
      FROM users
      WHERE username = 'admin_user'
    ),
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    )
  ),
  (
    'file2.jpg',
    '/path/to/file2.jpg',
    67890,
    'image/jpeg',
    (
      SELECT user_id
      FROM users
      WHERE username = 'regular_user'
    ),
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    )
  );
-- Stories
INSERT INTO stories (
    sequence_id,
    team_id,
    title,
    description,
    description_html,
    objective_id,
    status_id,
    assignee_id,
    reporter_id,
    priority,
    sprint_id,
    workspace_id,
    start_date,
    end_date,
    estimate,
    is_draft
  )
VALUES (
    1,
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    'Story 1',
    'Description of story 1',
    '<p>Description of story 1</p>',
    (
      SELECT objective_id
      FROM objectives
      WHERE name = 'Objective 1'
    ),
    (
      SELECT status_id
      FROM statuses
      WHERE name = 'Backlog'
        AND team_id = (
          SELECT team_id
          FROM teams
          WHERE name = 'Team Alpha'
        )
    ),
    (
      SELECT user_id
      FROM users
      WHERE username = 'admin_user'
    ),
    (
      SELECT user_id
      FROM users
      WHERE username = 'admin_user'
    ),
    'High',
    (
      SELECT sprint_id
      FROM sprints
      WHERE name = 'Sprint 1'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    ),
    '2024-01-01',
    '2024-01-05',
    3.0,
    false
  ),
  (
    2,
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    'Story 2',
    'Description of story 2',
    '<p>Description of story 2</p>',
    (
      SELECT objective_id
      FROM objectives
      WHERE name = 'Objective 2'
    ),
    (
      SELECT status_id
      FROM statuses
      WHERE name = 'Unstarted'
        AND team_id = (
          SELECT team_id
          FROM teams
          WHERE name = 'Team Beta'
        )
    ),
    (
      SELECT user_id
      FROM users
      WHERE username = 'regular_user'
    ),
    (
      SELECT user_id
      FROM users
      WHERE username = 'regular_user'
    ),
    'Medium',
    (
      SELECT sprint_id
      FROM sprints
      WHERE name = 'Sprint 2'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    ),
    '2024-02-01',
    '2024-02-05',
    5.0,
    true
  );
-- Labels
INSERT INTO labels (name, project_id, team_id, workspace_id, color)
VALUES (
    'Bug',
    (
      SELECT objective_id
      FROM objectives
      WHERE name = 'Objective 1'
    ),
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 1'
    ),
    '#FF0000'
  ),
  (
    'Feature',
    (
      SELECT objective_id
      FROM objectives
      WHERE name = 'Objective 2'
    ),
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    (
      SELECT workspace_id
      FROM workspaces
      WHERE name = 'Workspace 2'
    ),
    '#00FF00'
  );
-- Story Labels
INSERT INTO story_labels (story_id, label_id)
VALUES (
    (
      SELECT id
      FROM stories
      WHERE title = 'Story 1'
    ),
    (
      SELECT label_id
      FROM labels
      WHERE name = 'Bug'
    )
  ),
  (
    (
      SELECT id
      FROM stories
      WHERE title = 'Story 2'
    ),
    (
      SELECT label_id
      FROM labels
      WHERE name = 'Feature'
    )
  );
-- Team Members
INSERT INTO team_members (team_id, user_id, role)
VALUES (
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Alpha'
    ),
    (
      SELECT user_id
      FROM users
      WHERE username = 'admin_user'
    ),
    'Lead'
  ),
  (
    (
      SELECT team_id
      FROM teams
      WHERE name = 'Team Beta'
    ),
    (
      SELECT user_id
      FROM users
      WHERE username = 'regular_user'
    ),
    'Member'
  );