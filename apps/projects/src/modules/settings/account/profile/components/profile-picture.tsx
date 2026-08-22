"use client";

import { Avatar, ProfileUploadDialog, Tooltip } from "ui";
import { EditIcon } from "icons";
import { useState } from "react";
import { toast } from "sonner";
import type { User } from "@/types";
import { useProfile } from "@/lib/hooks/profile";
import { useUploadProfileImageMutation } from "@/lib/hooks/user/upload-profile-image-mutation";
import { useDeleteProfileImageMutation } from "@/lib/hooks/user/delete-profile-image-mutation";

export const ProfilePicture = ({
  initialProfile,
}: {
  initialProfile?: User;
}) => {
  const { data: profile } = useProfile(initialProfile);

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isUploading, setIsUploading] = useState(false);

  const { mutate: uploadProfileImage } = useUploadProfileImageMutation();
  const { mutate: deleteProfileImage } = useDeleteProfileImageMutation();

  const handleUpload = (file: File) => {
    setIsUploading(true);
    uploadProfileImage(file, {
      onSuccess: () => {
        setIsDialogOpen(false);
        setIsUploading(false);
        toast.info("Upload complete", {
          description: "Your profile image has been updated",
        });
      },
      onError: () => {
        setIsUploading(false);
      },
    });
  };

  const handleRemove = () => {
    deleteProfileImage(undefined, {
      onSuccess: () => {
        setIsDialogOpen(false);
        toast.info("Profile image removed", {
          description: "Your profile image has been removed",
        });
      },
    });
  };

  return (
    <>
      <Tooltip title="Change profile picture">
        <button
          aria-label="Change profile picture"
          className="group focus-visible:ring-primary/40 relative rounded-full outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
          onClick={() => {
            setIsDialogOpen(true);
          }}
          type="button"
        >
          <Avatar
            className="border-border group-hover:border-primary/50 h-12 border-2 transition-colors"
            name={profile?.fullName || profile?.username}
            src={profile?.avatarUrl}
          />
          <span
            aria-hidden="true"
            className="border-background bg-foreground text-background absolute -right-0.5 -bottom-0.5 flex size-5 items-center justify-center rounded-full border-2 transition-transform group-hover:scale-105"
          >
            <EditIcon className="h-2.5 w-auto" strokeWidth={2.5} />
          </span>
        </button>
      </Tooltip>
      <ProfileUploadDialog
        currentImage={profile?.avatarUrl}
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
