"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { parseAsString, useQueryStates } from "nuqs";
import { Button } from "ui";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import type { SaveViewProps } from "@/shared/views/save-slot";
import { ViewCreatorForm } from "./view-creator";
import { requestCreatedViewFocus } from "./view-selection";
import { savedViewPath } from "./view-link";

export const SaveViewAction = ({ teamId, configuration }: SaveViewProps) => {
  const { userRole } = useUserRole();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const { data: session } = useSession();
  const router = useRouter();
  const sourceScope = configuration.scope
    ? JSON.stringify(configuration.scope)
    : "team";
  const scope = `${workspaceSlug}:${session?.user.id ?? ""}:${sourceScope}:${teamId ?? ""}`;
  const [openingScope, setOpeningScope] = useState<string | null>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const [, setRoute] = useQueryStates({
    createView: parseAsString,
    visibility: parseAsString,
  });
  if (!userRole || userRole === "guest") return null;
  return (
    <>
      <Button
        color="tertiary"
        data-save-view-scope={scope}
        data-view-create-trigger={`${workspaceSlug}:${session?.user.id ?? ""}:${teamId ?? ""}`}
        onClick={() => {
          if (configuration.scope?.kind === "my-work") setOpeningScope(scope);
          else if (teamId)
            void setRoute({ createView: "true", visibility: null });
        }}
        ref={trigger}
        size="sm"
        variant="outline"
      >
        Save as
      </Button>
      {openingScope === scope ? (
        <ViewCreatorForm
          configuration={configuration}
          initialVisibility="personal"
          key={scope}
          onClose={async () => {
            setOpeningScope(null);
          }}
          onCreated={async (id, ownerTeamId) => {
            requestCreatedViewFocus({
              workspaceSlug,
              userId: session?.user.id ?? "",
              teamId: ownerTeamId,
              id,
            });
            setOpeningScope(null);
            router.push(withWorkspace(savedViewPath(ownerTeamId, id)));
          }}
          onReturnFocus={() => {
            if (trigger.current?.dataset.saveViewScope === scope)
              trigger.current.focus();
          }}
          teamId={teamId}
        />
      ) : null}
    </>
  );
};
