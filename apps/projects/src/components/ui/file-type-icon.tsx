import { cn } from "lib";
import { FLUENT_FILE_PATHS } from "./file-type-icons/fluent-paths";
import {
  ExcelFileGlyph,
  PowerPointFileGlyph,
  WordFileGlyph,
} from "./file-type-icons/office-glyphs";

export type FileIconKind =
  | "pdf"
  | "word"
  | "excel"
  | "powerpoint"
  | "csv"
  | "text"
  | "image"
  | "video"
  | "audio"
  | "archive"
  | "code"
  | "link"
  | "file";

const TILE_COLORS: Record<FileIconKind, string> = {
  pdf: "bg-[#E9155C]",
  word: "bg-[#EAF1FC] dark:bg-[#223650]",
  excel: "bg-[#E6F3EA] dark:bg-[#203D31]",
  powerpoint: "bg-[#FCF0E9] dark:bg-[#483129]",
  csv: "bg-[#168354]",
  text: "bg-[#657287]",
  image: "bg-[#8462BC]",
  video: "bg-[#8256BA]",
  audio: "bg-[#168EB5]",
  archive: "bg-[#AF7B2B]",
  code: "bg-[#5368AD]",
  link: "bg-[#2364C4]",
  file: "bg-[#73808C]",
};

const FileGlyph = ({ kind }: { kind: FileIconKind }) => {
  if (kind === "word") return <WordFileGlyph />;
  if (kind === "excel") return <ExcelFileGlyph />;
  if (kind === "powerpoint") return <PowerPointFileGlyph />;
  if (kind === "pdf") {
    return (
      <svg
        className="size-full"
        fill="none"
        focusable="false"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2.4}
        viewBox="0 0 48 48"
      >
        <path d="M24.6 13C24 10.6 21.4 11 21.6 14.2C21.9 20 25.9 26.5 31.4 28.2C36.1 29.7 36.1 25.5 31.9 25.5C25.6 25.3 18.2 27.3 14.6 31.6C12 34.8 15.2 36.4 17.7 32.7C21.5 27.1 24.6 18.1 24.6 13Z" />
      </svg>
    );
  }
  return (
    <svg
      className="size-3/5"
      fill="currentColor"
      focusable="false"
      viewBox="0 0 24 24"
    >
      {FLUENT_FILE_PATHS[kind].map((path) => (
        <path d={path} key={path} />
      ))}
    </svg>
  );
};

// The app owns the tile shape; vendored SVGs only supply recognizable glyphs.
// Asset sources, adaptations, and licenses: ./file-type-icons/NOTICE.md.
export const FileTypeIcon = ({
  kind,
  className,
}: {
  kind: FileIconKind;
  className?: string;
}) => (
  <span
    aria-hidden="true"
    className={cn(
      "inline-flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-md text-white",
      TILE_COLORS[kind],
      className,
    )}
  >
    <FileGlyph kind={kind} />
  </span>
);
