import Image from "next/image";
import { DocsIcon, ImageIcon, LayoutTable01Icon, Video02Icon } from "icons";
import { cn } from "lib";
import { getCommentFileType } from "./comment-file-type";

const BRAND_ASSETS = {
  word: "/images/file-types/docx.svg",
  excel: "/images/file-types/xlsx.svg",
  powerpoint: "/images/file-types/pptx.svg",
  pdf: "/images/file-types/pdf.svg",
};

const NEUTRAL_ICONS = {
  csv: LayoutTable01Icon,
  video: Video02Icon,
  image: ImageIcon,
  text: DocsIcon,
  file: DocsIcon,
};

export const CommentFileIcon = ({
  filename,
  mimeType,
  className,
}: {
  filename?: string;
  mimeType: string;
  className?: string;
}) => {
  const type = getCommentFileType(mimeType, filename);
  const classes = cn("size-10 shrink-0 object-contain", className);

  if (
    type === "word" ||
    type === "excel" ||
    type === "powerpoint" ||
    type === "pdf"
  ) {
    return (
      <Image
        alt=""
        aria-hidden="true"
        className={classes}
        draggable={false}
        height={48}
        src={BRAND_ASSETS[type]}
        unoptimized
        width={48}
      />
    );
  }

  const Icon = NEUTRAL_ICONS[type];

  return <Icon aria-hidden="true" className={classes} strokeWidth={1.5} />;
};
