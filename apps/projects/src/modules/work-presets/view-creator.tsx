"use client";

import { useEffect, useRef, useState } from "react";
import { parseAsString, useQueryStates } from "nuqs";
import { Button, Dialog, Input, Text } from "ui";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import type { ViewIconKey } from "@/shared/views/metadata";
import { usePresetMutations } from "./hooks";
import { ViewOwnerField } from "./view-owner-field";
import { ViewIconPicker } from "./view-icon-picker";
import { requestCreatedViewFocus } from "./view-selection";
import type { SavedViewConfiguration } from "./types";

export const ViewCreatorForm = ({
  teamId: initialTeamId,
  configuration,
  initialVisibility,
  onClose,
  onCreated,
  onReturnFocus,
}: {
  teamId?: string;
  configuration: SavedViewConfiguration;
  initialVisibility: "personal" | "team";
  onClose: () => Promise<void>;
  onCreated: (id: string, teamId: string) => Promise<void>;
  onReturnFocus: () => void;
}) => {
  const [teamId, setTeamId] = useState(initialTeamId ?? "");
  const [name, setName] = useState("");
  const [iconOverride, setIconOverride] = useState<
    ViewIconKey | null | undefined
  >();
  const icon =
    iconOverride === undefined ? configuration.icon ?? null : iconOverride;
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const mounted = useRef(true);
  const nameInput = useRef<HTMLInputElement>(null);
  const { userRole } = useUserRole();
  const { create } = usePresetMutations(teamId, "view");
  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);
  useEffect(() => {
    if (error && !pending) nameInput.current?.focus();
  }, [error, pending]);
  const canCreate = Boolean(userRole && userRole !== "guest");
  const save = async () => {
    if (pending || !canCreate || !teamId) return;
    const trimmed = name.trim();
    if (!trimmed || Array.from(trimmed).length > 100) {
      setError("Use a view name with 1–100 characters.");
      return;
    }
    setPending(true);
    setError(null);
    try {
      const saved = await create.mutateAsync({
        teamId,
        kind: "view",
        visibility: initialVisibility,
        name: trimmed,
        configuration: {
          ...configuration,
          ...(iconOverride !== undefined ? { icon: iconOverride } : {}),
        },
      });
      if (mounted.current) await onCreated(saved.id, teamId);
    } catch (cause) {
      if (mounted.current)
        setError(
          cause instanceof Error
            ? cause.message
            : "The view could not be created. Please try again.",
        );
    }
    if (mounted.current) setPending(false);
  };
  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open && !pending) void onClose();
      }}
      open
    >
      <Dialog.Content
        aria-busy={pending}
        className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
        hideClose={pending}
        onCloseAutoFocus={(event) => {
          event.preventDefault();
          onReturnFocus();
        }}
        onEscapeKeyDown={(event) => {
          if (pending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (pending) event.preventDefault();
        }}
        size="md"
      >
        <Dialog.Header className="shrink-0 px-6 py-4">
          <Dialog.Title className="text-lg">Create view</Dialog.Title>
          <Dialog.Description className="sr-only">
            Save the current filters, layout and display options.
          </Dialog.Description>
        </Dialog.Header>
        <form
          className="flex min-h-0 flex-1 flex-col"
          onSubmit={(event) => {
            event.preventDefault();
            void save();
          }}
        >
          <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-3">
            <ViewIconPicker
              disabled={pending || !canCreate}
              onChange={setIconOverride}
              value={icon}
            />
            <Input
              aria-label="View name"
              autoFocus
              disabled={pending || !canCreate}
              onChange={(event) => {
                setName(event.target.value);
              }}
              placeholder="View name"
              ref={nameInput}
              value={name}
            />
            {configuration.scope?.kind === "my-work" ? (
              <ViewOwnerField
                disabled={pending || !canCreate}
                onChange={setTeamId}
                value={teamId}
              />
            ) : null}
            {error ? (
              <Text color="danger" role="alert">
                {error}
              </Text>
            ) : null}
            {userRole === "guest" ? (
              <Text color="muted">
                Guests can use saved views but cannot create them.
              </Text>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="shrink-0 gap-2" justify="end">
            <Button
              color="tertiary"
              disabled={pending}
              onClick={() => void onClose()}
              type="button"
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              disabled={pending || !canCreate || !name.trim() || !teamId}
              loading={pending}
              loadingText="Creating..."
              type="submit"
            >
              Create view
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};

export const ViewCreator = ({
  teamId,
  configuration,
}: {
  teamId: string;
  configuration: SavedViewConfiguration;
}) => {
  const [route, setRoute] = useQueryStates({
    createView: parseAsString,
    view: parseAsString,
    visibility: parseAsString,
  });
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  if (route.createView !== "true") return null;
  const scope = `${workspaceSlug}:${session?.user.id ?? ""}:${teamId}`;
  const returnFocus = () => {
    requestAnimationFrame(() => {
      // A newer dialog owns focus if one opened while close animation finished.
      if (
        document.querySelector(
          '[role="dialog"][data-state="open"], [role="alertdialog"][data-state="open"]',
        )
      )
        return;
      const saveTrigger = Array.from(
        document.querySelectorAll<HTMLElement>("[data-view-create-trigger]"),
      ).find((element) => element.dataset.viewCreateTrigger === scope);
      if (saveTrigger) {
        saveTrigger.focus();
        return;
      }
      const header = document.querySelector<HTMLElement>(
        "[data-view-focus-scope]",
      );
      if (header?.dataset.viewFocusScope !== scope) return;
      header.querySelector<HTMLElement>("[data-view-menu-trigger]")?.focus();
    });
  };
  return (
    <ViewCreatorForm
      configuration={configuration}
      initialVisibility={route.visibility === "team" ? "team" : "personal"}
      key={`${scope}:${route.visibility ?? "personal"}`}
      onClose={async () => {
        await setRoute({ createView: null, visibility: null });
      }}
      onCreated={async (id) => {
        requestCreatedViewFocus({
          workspaceSlug,
          userId: session?.user.id ?? "",
          teamId,
          id,
        });
        await setRoute({ createView: null, visibility: null, view: id });
      }}
      onReturnFocus={returnFocus}
      teamId={teamId}
    />
  );
};
