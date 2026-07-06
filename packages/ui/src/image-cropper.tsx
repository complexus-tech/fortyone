"use client";

import { useCallback, useState } from "react";
import Cropper from "react-easy-crop";
import type { Area, Point } from "react-easy-crop";
import { CloseIcon } from "icons";
import { Button } from "./button";
import { Dialog } from "./dialog";

export type ImageCropperProps = {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  imageSrc: string;
  onCropComplete: (croppedAreaPixels: Area) => Promise<void> | void;
  aspectRatio?: number;
};

export const ImageCropper = ({
  isOpen,
  onOpenChange,
  imageSrc,
  onCropComplete,
  aspectRatio = 1,
}: ImageCropperProps) => {
  const [crop, setCrop] = useState<Point>({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<Area | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const handleCropComplete = useCallback(
    (_croppedArea: Area, nextCroppedAreaPixels: Area) => {
      setCroppedAreaPixels(nextCroppedAreaPixels);
    },
    [],
  );

  const handleSave = async () => {
    if (!croppedAreaPixels) {
      onOpenChange(false);
      return;
    }

    setIsSaving(true);
    try {
      await onCropComplete(croppedAreaPixels);
      onOpenChange(false);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content className="p-0" hideClose>
        <Dialog.Header className="flex items-center justify-between px-5 py-3">
          <Dialog.Title className="text-lg">Crop Image</Dialog.Title>
          <Button
            variant="naked"
            color="tertiary"
            size="xs"
            asIcon
            onClick={() => onOpenChange(false)}
          >
            <CloseIcon className="h-4.5" />
            <span className="sr-only">Close</span>
          </Button>
        </Dialog.Header>
        <Dialog.Body className="relative h-[45dvh] bg-black p-0">
          <Cropper
            image={imageSrc}
            crop={crop}
            zoom={zoom}
            aspect={aspectRatio}
            onCropChange={setCrop}
            onZoomChange={setZoom}
            onCropComplete={handleCropComplete}
          />
        </Dialog.Body>
        <Dialog.Footer
          className="gap-5 px-5 py-3"
          justify="end"
          variant="bordered"
        >
          <Button
            variant="naked"
            color="tertiary"
            className="px-5"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button onClick={handleSave} className="px-4" loading={isSaving}>
            Crop image
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

export type { Area as ImageCropArea };
