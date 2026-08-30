"use client";

import { useState } from "react";
import { ProfileUploadDialog } from "ui";
import { toast } from "sonner";
import { EditableAvatar } from "@/components/ui/editable-avatar";
import { useProfile } from "@/lib/hooks/profile";
import { useDeleteProfileImageMutation } from "@/lib/hooks/user/delete-profile-image-mutation";
import { useUploadProfileImageMutation } from "@/lib/hooks/user/upload-profile-image-mutation";
import type { User } from "@/types/user";

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
