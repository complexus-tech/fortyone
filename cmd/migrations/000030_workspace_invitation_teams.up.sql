-- 000030_workspace_invitation_teams.up.sql
CREATE TABLE "public"."workspace_invitation_teams" (
    "invitation_id" uuid NOT NULL,
    "team_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT "workspace_invitation_teams_team_id_fkey" FOREIGN KEY ("team_id") REFERENCES "public"."teams"("team_id") ON DELETE CASCADE,
    CONSTRAINT "workspace_invitation_teams_invitation_id_fkey" FOREIGN KEY ("invitation_id") REFERENCES "public"."workspace_invitations"("invitation_id") ON DELETE CASCADE,
    PRIMARY KEY ("invitation_id","team_id")
);
