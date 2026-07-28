export const getAdjacentAttachmentIndex = (
  currentIndex: number,
  attachmentCount: number,
  direction: -1 | 1,
) => {
  if (attachmentCount <= 0) return 0;
  return (currentIndex + direction + attachmentCount) % attachmentCount;
};
