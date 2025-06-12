import { Avatar } from "ui";
import { useState } from "react";
import { ProfileUploadDialog } from "ui/src/ProfileUploadDialog/ProfileUploadDialog";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { uploadWorkspaceLogoAction } from "@/lib/actions/workspaces/upload-logo";
import { useUpdateWorkspaceMutation } from "@/lib/hooks/update-workspace-mutation";
import { workspaceKeys } from "@/constants/keys";

export const Logo = () => {
  const { workspace } = useCurrentWorkspace();
  const [previewUrl, setPreviewUrl] = useState<string | undefined>(
    workspace?.avatarUrl,
  );
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const { mutateAsync: updateWorkspace } = useUpdateWorkspaceMutation();
  const queryClient = useQueryClient();

  const handleUpload = async (file: File) => {
    setIsUploading(true);
    const res = await uploadWorkspaceLogoAction(file);
    if (res.error?.message) {
      toast.error("Failed to upload logo", {
        description:
          res.error.message || "An error occurred while uploading the file",
        action: {
          label: "Retry",
          onClick: () => {
            handleUpload(file);
          },
        },
      });
      setIsUploading(false);
      return;
    }
    toast.info("Upload complete", {
      description: "Your workspace logo has been updated",
    });
    setIsUploading(false);
    setIsDialogOpen(false);
    queryClient.invalidateQueries({ queryKey: workspaceKeys.lists() });
  };

  const handleRemove = () => {
    updateWorkspace({ avatarUrl: "" });
    setPreviewUrl(undefined);
  };
  return (
    <>
      <button
        onClick={() => {
          setIsDialogOpen(true);
        }}
        type="button"
      >
        <Avatar
          className="h-10"
          name={workspace?.name}
          src={workspace?.avatarUrl}
        />
        <span className="sr-only">Change workspace logo</span>
      </button>
      <ProfileUploadDialog
        currentImage={previewUrl}
        isOpen={isDialogOpen}
        isUploading={isUploading}
        maxSizeInMB={5}
        onOpenChange={setIsDialogOpen}
        onRemove={handleRemove}
        onUpload={handleUpload}
      />
    </>
  );
};
