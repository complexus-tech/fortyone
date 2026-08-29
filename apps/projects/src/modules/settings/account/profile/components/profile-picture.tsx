"use client";

import { ProfileUploadDialog } from "ui";
import { useState } from "react";
import { toast } from "sonner";
import type { User } from "@/types";
import { useProfile } from "@/lib/hooks/profile";
import { useUploadProfileImageMutation } from "@/lib/hooks/user/upload-profile-image-mutation";
import { useDeleteProfileImageMutation } from "@/lib/hooks/user/delete-profile-image-mutation";
import { EditableAvatar } from "@/modules/settings/components";

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
      <EditableAvatar
        label="Change profile picture"
        name={profile?.fullName || profile?.username}
        onClick={() => {
          setIsDialogOpen(true);
        }}
        src={profile?.avatarUrl}
      />
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
