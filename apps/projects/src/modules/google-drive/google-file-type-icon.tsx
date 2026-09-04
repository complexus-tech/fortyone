import { DocsIcon, ImageIcon } from "icons";
import { cn } from "lib";
import {
  GoogleDocsIcon,
  GoogleSheetsIcon,
} from "./google-workspace-file-icons";

const GOOGLE_DOCS_MIME_TYPE = "application/vnd.google-apps.document";
const GOOGLE_SHEETS_MIME_TYPE = "application/vnd.google-apps.spreadsheet";

export const isGoogleDocsMimeType = (mimeType: string) =>
  mimeType === GOOGLE_DOCS_MIME_TYPE;

export const isGoogleSheetsMimeType = (mimeType: string) =>
  mimeType === GOOGLE_SHEETS_MIME_TYPE;

export const hasNativeGoogleWorkspaceIcon = (mimeType: string) =>
  isGoogleDocsMimeType(mimeType) || isGoogleSheetsMimeType(mimeType);

const getIconPresentation = (mimeType: string) => {
  if (mimeType.includes("presentation")) {
    return {
      className: "bg-[#fef7e0] text-[#f9ab00] dark:bg-[#f9ab00]/20",
      icon: DocsIcon,
    };
  }
  if (mimeType.startsWith("image/")) {
    return {
      className: "bg-[#f3e8fd] text-[#a142f4] dark:bg-[#a142f4]/20",
      icon: ImageIcon,
    };
  }
  if (mimeType === "application/pdf") {
    return {
      className: "bg-[#fce8e6] text-[#d93025] dark:bg-[#d93025]/20",
      icon: DocsIcon,
    };
  }
  return {
    className: "bg-[#e8f0fe] text-[#1967d2] dark:bg-[#1967d2]/20",
    icon: DocsIcon,
  };
};

export const GoogleFileTypeIcon = ({
  className,
  mimeType,
}: {
  className?: string;
  mimeType: string;
}) => {
  if (isGoogleDocsMimeType(mimeType) || isGoogleSheetsMimeType(mimeType)) {
    const Icon = isGoogleDocsMimeType(mimeType)
      ? GoogleDocsIcon
      : GoogleSheetsIcon;

    return (
      <span
        className={cn(
          "flex size-9 shrink-0 items-center justify-center",
          className,
        )}
      >
        <Icon className="size-[72%]" />
      </span>
    );
  }

  const presentation = getIconPresentation(mimeType);
  const Icon = presentation.icon;
  return (
    <span
      className={cn(
        "flex size-9 shrink-0 items-center justify-center rounded-lg",
        presentation.className,
        className,
      )}
    >
      <Icon aria-hidden="true" className="size-4.5" />
    </span>
  );
};

export const getGoogleFileTypeLabel = (mimeType: string) => {
  if (isGoogleSheetsMimeType(mimeType)) return "Google Sheet";
  if (mimeType.includes("presentation")) return "Google Slides";
  if (isGoogleDocsMimeType(mimeType)) return "Google Doc";
  if (mimeType === "application/pdf") return "PDF";
  if (mimeType.startsWith("image/")) return "Image";
  return "Google Drive file";
};
