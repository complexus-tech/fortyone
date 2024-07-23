-- User Management
CREATE TABLE roles (
  role_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) UNIQUE NOT NULL,
  permissions JSONB NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(255) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  full_name VARCHAR(255) NOT NULL,
  role_id UUID REFERENCES roles(role_id) NOT NULL,
  avatar_url VARCHAR(255),
  is_active BOOLEAN DEFAULT true NOT NULL,
  last_login_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- Team Management
CREATE TABLE teams (
  team_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- Workspace Management
CREATE TABLE workspaces (
  workspace_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- Project Management
CREATE TABLE objectives (
  objective_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) UNIQUE NOT NULL,
  description TEXT,
  lead_user_id UUID REFERENCES users(user_id),
  team_id UUID REFERENCES teams(team_id),
  workspace_id UUID REFERENCES workspaces(workspace_id),
  start_date DATE,
  end_date DATE,
  status VARCHAR(255),
  is_private BOOLEAN DEFAULT false NOT NULL,
  visibility VARCHAR(255),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE statuses (
  status_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  color VARCHAR(255),
  category VARCHAR(255),
  order_index INTEGER,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE sprints (
  sprint_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  objective_id UUID REFERENCES objectives(objective_id) NOT NULL,
  team_id UUID REFERENCES teams(team_id),
  workspace_id UUID REFERENCES workspaces(workspace_id),
  name VARCHAR(255) NOT NULL,
  start_date DATE,
  end_date DATE,
  goal TEXT,
  status VARCHAR(255),
  backlog_status VARCHAR(255),
  completed_story_count INTEGER DEFAULT 0 NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- File Management
CREATE TABLE attachments (
  attachment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  filename VARCHAR(255) NOT NULL,
  path VARCHAR(255) NOT NULL,
  size BIGINT NOT NULL,
  mime_type VARCHAR(255) NOT NULL,
  uploaded_by UUID REFERENCES users(user_id) NOT NULL,
  team_id UUID REFERENCES teams(team_id),
  workspace_id UUID REFERENCES workspaces(workspace_id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- Story Management
CREATE TABLE stories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  sequence_id SERIAL,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  description_html TEXT,
  parent_id UUID REFERENCES stories(id),
  objective_id UUID REFERENCES objectives(objective_id) NOT NULL,
  status_id UUID REFERENCES statuses(status_id) NOT NULL,
  assignee_id UUID REFERENCES users(user_id),
  blocked_by_id UUID REFERENCES stories(id),
  blocking_id UUID REFERENCES stories(id),
  related_id UUID REFERENCES stories(id),
  reporter_id UUID REFERENCES users(user_id) NOT NULL,
  priority VARCHAR(100),
  sprint_id UUID REFERENCES sprints(sprint_id),
  team_id UUID REFERENCES teams(team_id),
  workspace_id UUID REFERENCES workspaces(workspace_id),
  start_date DATE,
  end_date DATE,
  estimate FLOAT,
  archived_at TIMESTAMP WITH TIME ZONE,
  is_draft BOOLEAN DEFAULT false NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE TABLE story_attachments (
  story_id UUID REFERENCES stories(id) NOT NULL,
  attachment_id UUID REFERENCES attachments(attachment_id) NOT NULL,
  PRIMARY KEY (story_id, attachment_id)
);
CREATE TABLE story_blockers (
  blocker_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  blocked_by_id UUID REFERENCES stories(id) NOT NULL,
  block_id UUID REFERENCES stories(id) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_watchers (
  story_id UUID REFERENCES stories(id) NOT NULL,
  user_id UUID REFERENCES users(user_id) NOT NULL,
  PRIMARY KEY (story_id, user_id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_relations (
  relation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  story_id UUID REFERENCES stories(id) NOT NULL,
  related_story_id UUID REFERENCES stories(id) NOT NULL,
  relation_type VARCHAR(20) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_mentions (
  mention_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  story_id UUID REFERENCES stories(id) NOT NULL,
  mentioner_id UUID REFERENCES users(user_id) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_assignees (
  assignee_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  story_id UUID REFERENCES stories(id) NOT NULL,
  user_id UUID REFERENCES users(user_id) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_links (
  link_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR(255) NOT NULL,
  url VARCHAR(255) NOT NULL,
  story_id UUID REFERENCES stories(id) NOT NULL,
  metadata JSONB,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_comments (
  comment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content TEXT NOT NULL,
  story_id UUID REFERENCES stories(id) NOT NULL,
  commenter_id UUID REFERENCES users(user_id) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_activities (
  activity_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  story_id UUID REFERENCES stories(id) NOT NULL,
  sprint_id UUID REFERENCES sprints(sprint_id) NOT NULL,
  type VARCHAR(255) NOT NULL,
  description TEXT,
  estimate INTEGER,
  assignee_id UUID REFERENCES users(user_id),
  start_date TIMESTAMP WITH TIME ZONE,
  end_date TIMESTAMP WITH TIME ZONE,
  status VARCHAR(255),
  progress INTEGER,
  dependencies VARCHAR(255),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE labels (
  label_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  project_id UUID REFERENCES objectives(objective_id) NOT NULL,
  team_id UUID REFERENCES teams(team_id),
  workspace_id UUID REFERENCES workspaces(workspace_id),
  color VARCHAR(7),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_labels (
  story_id UUID REFERENCES stories(id) NOT NULL,
  label_id UUID REFERENCES labels(label_id) NOT NULL,
  PRIMARY KEY (story_id, label_id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE custom_fields (
  field_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  type VARCHAR(255) NOT NULL,
  project_id UUID REFERENCES objectives(objective_id) NOT NULL,
  team_id UUID REFERENCES teams(team_id),
  workspace_id UUID REFERENCES workspaces(workspace_id),
  options JSONB,
  is_required BOOLEAN DEFAULT false NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_custom_fields (
  story_id UUID REFERENCES stories(id) NOT NULL,
  field_id UUID REFERENCES custom_fields(field_id) NOT NULL,
  value TEXT,
  PRIMARY KEY (story_id, field_id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE story_votes (
  vote_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  story_id UUID REFERENCES stories(id) NOT NULL,
  voter_id UUID REFERENCES users(user_id) NOT NULL,
  vote INTEGER NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE team_members (
  team_id UUID REFERENCES teams(team_id) NOT NULL,
  user_id UUID REFERENCES users(user_id) NOT NULL,
  role VARCHAR(255),
  PRIMARY KEY (team_id, user_id),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE notifications (
  notification_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(user_id) NOT NULL,
  content TEXT NOT NULL,
  type VARCHAR(255),
  is_read BOOLEAN DEFAULT false NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE integrations (
  integration_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  type VARCHAR(255) NOT NULL,
  config JSONB,
  user_id UUID REFERENCES users(user_id) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- Indexes for performance optimization
CREATE INDEX idx_stories_objective_id ON stories(objective_id);
CREATE INDEX idx_stories_sprint_id ON stories(sprint_id);
CREATE INDEX idx_stories_team_id ON stories(team_id);
CREATE INDEX idx_stories_workspace_id ON stories(workspace_id);
CREATE INDEX idx_objectives_lead_user_id ON objectives(lead_user_id);
CREATE INDEX idx_attachments_uploaded_by ON attachments(uploaded_by);