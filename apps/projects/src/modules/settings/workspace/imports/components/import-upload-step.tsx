"use client";
import { useDropzone } from "react-dropzone";
import { DownloadIcon } from "icons";
import { Box, DropZone, Flex, Text } from "ui";
import { IMPORT_MAX_FILE_BYTES } from "../schema";
import { ImportAnalysisBanner } from "./import-analysis-banner";

const fileAccept = {
  "text/csv": [".csv"],
  "text/tab-separated-values": [".tsv"],
  "application/json": [".json"],
  "application/vnd.ms-excel": [".xls"],
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
    ".xlsx",
  ],
  "application/pdf": [".pdf"],
  "image/jpeg": [".jpg", ".jpeg"],
  "image/png": [".png"],
  "image/webp": [".webp"],
};

type ImportUploadStepProps = {
  analysisError: string;
  analysisNotice: string;
  analysisPending: boolean;
  fileName: string;
  hasAttemptedImport: boolean;
  uploadPending: boolean;
  handleFile: (file: File) => void;
  setAnalysisError: (message: string) => void;
};
export const ImportUploadStep = ({
  analysisError,
  analysisNotice,
  analysisPending,
  fileName,
  hasAttemptedImport,
  uploadPending,
  handleFile,
  setAnalysisError,
}: ImportUploadStepProps) => {
  const dropzone = useDropzone({
    accept: fileAccept,
    disabled: uploadPending || analysisPending,
    maxFiles: 1,
    maxSize: IMPORT_MAX_FILE_BYTES,
    multiple: false,
    onDropAccepted: ([file]) => {
      handleFile(file);
    },
    onDropRejected: (rejections) => {
      const tooLarge = rejections.some(({ errors }) =>
        errors.some((error) => error.code === "file-too-large"),
      );
      let message =
        "Choose a CSV, TSV, JSON, Excel, PDF, JPG, PNG, or WebP file.";
      if (tooLarge) message = "The import file must be 20 MB or smaller.";
      setAnalysisError(message);
    },
  });

  let uploadLabel = "Drop a file here or choose one";
  if (uploadPending) uploadLabel = "Reading your file…";
  else if (dropzone.isDragActive) uploadLabel = "Drop it here";

  return (
    <Box>
      <Text as="h2" className="text-xl font-medium">
        Add your export
      </Text>
      <Text className="mt-1 leading-6" color="muted">
        Export projects, teams, tasks, or issues from Jira, Trello, ClickUp,
        monday.com, Asana, or another tool, then upload the file here. Mapping
        starts automatically, and you can review every suggestion before the
        import runs.
      </Text>

      {fileName ? (
        <DropZone>
          <DropZone.Input inputProps={dropzone.getInputProps()} />
          <ImportAnalysisBanner
            analysisError={analysisError}
            analysisNotice={analysisNotice}
            analysisPending={analysisPending}
            fileName={fileName}
            hasAttemptedImport={hasAttemptedImport}
            key={fileName}
            onReplace={dropzone.open}
            uploadPending={uploadPending}
          />
        </DropZone>
      ) : (
        <DropZone>
          <DropZone.Root
            className="bg-surface-muted/35 mt-5 h-44"
            isDragActive={dropzone.isDragActive}
            rootProps={dropzone.getRootProps()}
          >
            <DropZone.Input inputProps={dropzone.getInputProps()} />
            <Flex align="center" direction="column" gap={2}>
              <Box className="bg-surface flex h-12 w-12 items-center justify-center rounded-xl">
                <DownloadIcon className="h-6" />
              </Box>
              <Text className="font-medium">{uploadLabel}</Text>
              <Text color="muted">
                CSV, JSON, Excel, PDF, JPG, PNG, or WebP, up to 20 MB
              </Text>
            </Flex>
          </DropZone.Root>
        </DropZone>
      )}

      {analysisError && !fileName ? (
        <Box className="bg-danger/8 mt-4 rounded-2xl p-4">
          <Text className="text-danger font-medium">Unable to continue</Text>
          <Text className="mt-1" color="muted">
            {analysisError}
          </Text>
        </Box>
      ) : null}
    </Box>
  );
};
