import { Avatar } from "ui";
import { useState } from "react";
import { ProfileUploadDialog } from "ui";
import { toast } from "sonner";
import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useUploadWorkspaceLogoMutation } from "@/lib/hooks/workspace/upload-logo-mutation";
import { useDeleteWorkspaceLogoMutation } from "@/lib/hooks/workspace/delete-logo-mutation";

export const Logo = () => {
  const { workspace } = useCurrentWorkspace();
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const { mutate: uploadLogo } = useUploadWorkspaceLogoMutation();
  const { mutate: deleteLogo } = useDeleteWorkspaceLogoMutation();
  const handleUpload = (file: File) => {
    uploadLogo(file, {
      onSuccess: () => {
        setIsDialogOpen(false);
        toast.info("Upload complete", {
          description: "Your workspace logo has been updated",
        });
      },
    });
  };

  const handleRemove = () => {
    deleteLogo(undefined, {
      onSuccess: () => {
        toast.info("Logo removed", {
          description: "Your workspace logo has been removed",
        });
      },
    });
    setIsDialogOpen(false);
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
        currentImage={workspace?.avatarUrl}
        isOpen={isDialogOpen}
        isUploading={workspace?.avatarUrl?.includes("blob:")}
        maxSizeInMB={5}
        onOpenChange={setIsDialogOpen}
        onRemove={handleRemove}
        onUpload={handleUpload}
      />
    </>
  );
};
