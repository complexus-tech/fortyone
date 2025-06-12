"use client";
import { useState, useCallback, useRef } from "react";
import { useDropzone } from "react-dropzone";
import { Dialog } from "../Dialog/Dialog";
import { Button } from "../Button/Button";
import { Text } from "../Text/Text";
import { BlurImage } from "../Image/Image";
import { DropZone } from "../DropZone/DropZone";
import { Flex } from "../Flex/Flex";
import { Box } from "../Box/Box";

interface ProfileUploadDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onUpload: (file: File) => void;
  onRemove: () => void;
  currentImage?: string;
  maxSizeInMB?: number;
  isUploading?: boolean;
  title?: string;
}

const SUPPORTED_FORMATS = [
  "image/jpeg",
  "image/jpg",
  "image/png",
  "image/webp",
];
const SUPPORTED_EXTENSIONS = [".jpeg", ".jpg", ".png", ".webp"];

export const ProfileUploadDialog = ({
  isOpen,
  onOpenChange,
  onUpload,
  onRemove,
  currentImage,
  isUploading,
  maxSizeInMB = 5,
  title = "Upload Image",
}: ProfileUploadDialogProps) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const validateFile = (file: File): string | null => {
    if (!SUPPORTED_FORMATS.includes(file.type)) {
      return `Unsupported file format. Please use: ${SUPPORTED_EXTENSIONS.join(", ")}`;
    }

    const maxSizeInBytes = maxSizeInMB * 1024 * 1024;
    if (file.size > maxSizeInBytes) {
      return `File size too large. Maximum size is ${maxSizeInMB}MB.`;
    }

    return null;
  };

  const handleFileSelect = useCallback(
    (file: File) => {
      const validationError = validateFile(file);
      if (validationError) {
        setError(validationError);
        return;
      }
      setError(null);
      setSelectedFile(file);
      const url = URL.createObjectURL(file);
      setPreviewUrl(url);
    },
    [maxSizeInMB]
  );

  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      if (acceptedFiles.length > 0) {
        handleFileSelect(acceptedFiles[0]);
      }
    },
    [handleFileSelect]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      "image/*": SUPPORTED_EXTENSIONS,
    },
    multiple: false,
    maxSize: maxSizeInMB * 1024 * 1024,
  });

  const handleEditClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileInputChange = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleUpload = () => {
    if (selectedFile) {
      onUpload(selectedFile);
      handleClose();
    }
  };

  const handleRemove = () => {
    onRemove();
    handleClose();
  };

  const handleClose = () => {
    setSelectedFile(null);
    setPreviewUrl(null);
    setError(null);
    onOpenChange(false);
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl);
    }
  };

  const displayImage = previewUrl || currentImage;
  const hasImage = Boolean(displayImage);

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content size="lg" className="max-w-2xl">
        <Dialog.Header className="px-6">
          <Dialog.Title className="text-xl">{title}</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="px-6 py-5">
          {hasImage ? (
            <Box className="relative">
              <Box className="relative mx-auto size-96 overflow-hidden rounded-2xl border-2 border-gray-200 dark:border-white/20">
                <BlurImage
                  src={displayImage!}
                  alt="Profile preview"
                  className="h-full w-full"
                  imageClassName="object-cover object-top"
                />
                <Button
                  onClick={handleEditClick}
                  color="white"
                  size="xs"
                  className="absolute right-2 top-2 rounded-lg tracking-wide"
                >
                  Change
                </Button>
              </Box>
              <input
                ref={fileInputRef}
                type="file"
                accept={SUPPORTED_EXTENSIONS.join(",")}
                onChange={handleFileInputChange}
                className="hidden"
              />
            </Box>
          ) : (
            <DropZone>
              <DropZone.Root
                rootProps={getRootProps()}
                isDragActive={isDragActive}
                className="h-96 w-96 mx-auto dark:border-white/20"
              >
                <DropZone.Input inputProps={getInputProps()} />
                <DropZone.Body
                  isDragActive={isDragActive}
                  message="Click or drag image here"
                />
              </DropZone.Root>
            </DropZone>
          )}
          <Text color="muted" className="mt-6">
            File formats supported: {SUPPORTED_EXTENSIONS.join(", ")}
          </Text>
          {error && (
            <Text color="danger" className="mt-2">
              {error}
            </Text>
          )}
        </Dialog.Body>
        <Dialog.Footer justify="between" className="gap-3">
          <Button
            color="tertiary"
            onClick={handleRemove}
            disabled={!currentImage}
          >
            Remove
          </Button>
          <Flex gap={3}>
            <Button variant="naked" color="tertiary" onClick={handleClose}>
              Cancel
            </Button>
            <Button
              onClick={handleUpload}
              disabled={!selectedFile}
              loading={isUploading}
              loadingText="Uploading..."
            >
              Upload & Save
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
