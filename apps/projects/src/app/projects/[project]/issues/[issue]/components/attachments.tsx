import { useDropzone } from "react-dropzone";
import { Box, DropZone, Text } from "ui";
import { AttachmentIcon } from "icons";
import { useCallback, useState } from "react";

interface CustomFile extends File {
  preview?: string;
}

export const Attachments = () => {
  const [files, setFiles] = useState<CustomFile[]>([]);

  const onDrop = useCallback((acceptedFiles: CustomFile[]) => {
    if (acceptedFiles.length > 0) {
      setFiles((prevFiles) => [
        ...prevFiles,
        ...acceptedFiles.map((file) =>
          Object.assign(file, {
            preview: URL.createObjectURL(file),
          }),
        ),
      ]);
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
  });

  return (
    <Box>
      <Text
        as="h4"
        className="mb-1 flex items-center gap-1"
        fontWeight="medium"
      >
        <AttachmentIcon className="h-5 w-auto" />
        Attachements
      </Text>
      <DropZone>
        <DropZone.Root isDragActive={isDragActive} rootProps={getRootProps()}>
          <DropZone.Input
            inputProps={getInputProps({
              multiple: true,
            })}
          />
          <DropZone.Body isDragActive={isDragActive} />
        </DropZone.Root>
      </DropZone>

      <Box className="mt-3 grid grid-cols-5 gap-3">
        {files.map((file, index) => (
          <Box
            className="relative h-24 overflow-hidden rounded-xl bg-gray-50/70 dark:bg-dark-200/50"
            key={index}
          >
            <Box className="absolute right-2 top-2">
              <Text className="text-xs font-medium text-white">X</Text>
            </Box>
            <Box className="absolute bottom-2 left-2">
              <Text className="text-xs font-medium text-white">
                {file.name}
              </Text>
            </Box>
          </Box>
        ))}
      </Box>
    </Box>
  );
};
