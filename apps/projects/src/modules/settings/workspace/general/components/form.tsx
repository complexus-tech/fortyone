import { Box, Input, Button } from "ui";
import type { ChangeEvent, FormEvent } from "react";
import { useEffect, useState, useMemo } from "react";
import { toast } from "sonner";
import { CopyIcon } from "icons";
import { useUpdateWorkspaceMutation } from "@/lib/hooks/update-workspace-mutation";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useCopyToClipboard } from "@/hooks";

export const WorkspaceForm = () => {
  const { mutateAsync: updateWorkspace } = useUpdateWorkspaceMutation();
  const { workspace } = useCurrentWorkspace();
  const [host, setHost] = useState("");
  const [_, copy] = useCopyToClipboard();
  const [form, setForm] = useState({
    name: workspace?.name || "",
  });

  const hasChanges = useMemo(() => {
    return workspace?.name !== form.name;
  }, [workspace, form]);

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    await updateWorkspace({ name: form.name });
    toast.success("Success", { description: "Workspace updated" });
  };

  useEffect(() => {
    if (workspace) {
      setForm({ name: workspace.name });
      setHost(window.location.hostname);
    }
  }, [workspace]);

  return (
    <form className="p-6" onSubmit={handleSubmit}>
      <Box className="mb-4 grid gap-4 md:grid-cols-2 md:gap-6">
        <Input
          label="Name"
          name="name"
          onChange={handleChange}
          placeholder="Enter workspace name"
          value={form.name}
        />
        <Input
          className="cursor-default lowercase opacity-80"
          label="URL (read-only)"
          name="slug"
          onClick={() => {
            copy(`https://${host}`);
            toast.success("Copied to clipboard");
          }}
          readOnly
          rightIcon={<CopyIcon />}
          value={`https://${host}`}
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
