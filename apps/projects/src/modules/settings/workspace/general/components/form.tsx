import { Box, Input, Button } from "ui";
import type { ChangeEvent, FormEvent } from "react";
import { useState } from "react";
import { toast } from "sonner";
import { CopyIcon } from "icons";
import { useUpdateWorkspaceMutation } from "@/lib/hooks/update-workspace-mutation";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useCopyToClipboard } from "@/hooks";

const isFortyOneApp = process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

export const WorkspaceForm = () => {
  const { mutateAsync: updateWorkspace } = useUpdateWorkspaceMutation();
  const { workspace } = useCurrentWorkspace();
  const origin = typeof window === "undefined" ? "" : window.location.origin;
  const [, copy] = useCopyToClipboard();
  const [name, setName] = useState<string | undefined>(undefined);
  const workspaceName = workspace?.name ?? "";
  const currentName = name ?? workspaceName;

  const getWorkspaceUrl = () => {
    if (isFortyOneApp) {
      return origin;
    }
    return `${origin}/${workspace?.slug}`;
  };

  const hasChanges = workspaceName !== currentName;

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value);
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    await updateWorkspace({ name: currentName.trim() });
    toast.success("Success", { description: "Workspace updated" });
  };

  return (
    <form className="p-6" onSubmit={handleSubmit}>
      <Box className="mb-4 grid gap-4 md:grid-cols-2 md:gap-6">
        <Input
          label="Name"
          name="name"
          onChange={handleChange}
          placeholder="Enter workspace name"
          value={currentName}
        />
        <Input
          className="cursor-default lowercase opacity-80"
          label="URL (read-only)"
          name="slug"
          onClick={() => {
            copy(getWorkspaceUrl());
            toast.success("Copied to clipboard");
          }}
          readOnly
          rightIcon={<CopyIcon />}
          value={getWorkspaceUrl()}
        />
      </Box>
      {hasChanges ? (
        <Button disabled={!hasChanges} type="submit">
          Save changes
        </Button>
      ) : null}
    </form>
  );
};
