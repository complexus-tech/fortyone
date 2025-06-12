"use client";

import { Avatar } from "ui";
import { ProfileUploadDialog } from "ui/src/ProfileUploadDialog/ProfileUploadDialog";
import { useState } from "react";
import { toast } from "sonner";
import { useProfile } from "@/lib/hooks/profile";
import { useUploadProfileImageMutation } from "@/lib/hooks/user/upload-profile-image-mutation";
import { useDeleteProfileImageMutation } from "@/lib/hooks/user/delete-profile-image-mutation";

export const ProfilePicture = () => {
  const { data: profile } = useProfile();

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
      <button
        onClick={() => {
          setIsDialogOpen(true);
        }}
        type="button"
      >
        <Avatar
          className="h-10"
          name={profile?.fullName || profile?.username}
          src={profile?.avatarUrl}
        />
        <span className="sr-only">Change profile picture</span>
      </button>
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
