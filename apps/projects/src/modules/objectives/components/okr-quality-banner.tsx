import { AiIcon, CheckIcon, WarningIcon } from "icons";
import { cn } from "lib";
import { Box, Button, Flex, Text } from "ui";
import type { OkrQualityAssessment } from "../schemas/okr-quality";

export const OkrQualityBanner = ({
  assessment,
  isAssessing,
  onUseSuggestion,
}: {
  assessment: OkrQualityAssessment | null;
  isAssessing: boolean;
  onUseSuggestion?: (suggestion: string) => void;
}) => {
  if (!assessment && !isAssessing) return null;

  const verdict = assessment?.verdict ?? "strong";
  const isStrong = verdict === "strong";
  const isDuplicate = verdict === "duplicate";
  let Icon = WarningIcon;
  let tone: "success" | "danger" | "warning" = "warning";
  if (isAssessing) Icon = AiIcon;
  if (isStrong) {
    Icon = isAssessing ? AiIcon : CheckIcon;
    tone = "success";
  } else if (isDuplicate) {
    tone = "danger";
  }
  const suggestedName = assessment?.suggestedName;

  return (
    <Flex
      align="start"
      className={cn("mt-3 rounded-xl border px-4 py-3", {
        "border-success/20 bg-success/5": tone === "success",
        "border-danger/20 bg-danger/5": tone === "danger",
        "border-warning/20 bg-warning/5": tone === "warning",
      })}
      gap={2}
      justify="between"
    >
      <Flex align="start" className="min-w-0" gap={2}>
        <Icon
          className={cn("mt-0.5 h-5 shrink-0", {
            "text-success dark:text-success": tone === "success",
            "text-danger dark:text-danger": tone === "danger",
            "text-warning dark:text-warning": tone === "warning",
          })}
        />
        <Box className="min-w-0">
          <Text fontWeight="medium">
            {isAssessing && !assessment
              ? "Maya is reviewing this goal…"
              : assessment?.headline}
          </Text>
          {assessment?.duplicateOf ? (
            <Text className="mt-0.5" color="muted">
              Similar to “{assessment.duplicateOf}”
            </Text>
          ) : null}
          {assessment?.guidance.map((item) => (
            <Text className="mt-0.5" color="muted" key={item}>
              {item}
            </Text>
          ))}
        </Box>
      </Flex>
      {suggestedName && onUseSuggestion ? (
        <Button
          className="shrink-0"
          color={tone === "success" ? "primary" : tone}
          onClick={() => {
            onUseSuggestion(suggestedName);
          }}
          size="xs"
          type="button"
          variant="naked"
        >
          Use suggestion
        </Button>
      ) : null}
    </Flex>
  );
};
