import { useEffect, useMemo, useRef, useState } from "react";
import { useDebouncedCallback } from "@/hooks/debounce";
import {
  getLocalOkrQualityAssessment,
  isReadyForAiQualityAssessment,
} from "../okr-quality";
import {
  okrQualityAssessmentSchema,
  okrQualityRequestSchema,
  type OkrQualityAssessment,
  type OkrQualityRequest,
} from "../schemas/okr-quality";

export const useOkrQualityAssessment = (request: OkrQualityRequest | null) => {
  const [assessment, setAssessment] = useState<OkrQualityAssessment | null>(
    null,
  );
  const [isAssessing, setIsAssessing] = useState(false);
  const controllerRef = useRef<AbortController | null>(null);

  const { callback: assess, cancel } = useDebouncedCallback(
    async (nextRequest: OkrQualityRequest) => {
      controllerRef.current?.abort();
      const controller = new AbortController();
      controllerRef.current = controller;
      setIsAssessing(true);

      try {
        const response = await fetch("/api/assess-okr-quality", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(nextRequest),
          signal: controller.signal,
        });
        if (!response.ok) throw new Error(await response.text());
        const parsed = okrQualityAssessmentSchema.safeParse(
          await response.json(),
        );
        if (parsed.success) setAssessment(parsed.data);
      } catch (error) {
        if (!(error instanceof DOMException && error.name === "AbortError")) {
          setAssessment(getLocalOkrQualityAssessment(nextRequest));
        }
      } finally {
        if (controllerRef.current === controller) {
          controllerRef.current = null;
          setIsAssessing(false);
        }
      }
    },
    900,
  );

  const requestKey = request ? JSON.stringify(request) : "";
  const stableRequest = useMemo(() => {
    if (!requestKey) return null;
    const parsed = okrQualityRequestSchema.safeParse(JSON.parse(requestKey));
    return parsed.success ? parsed.data : null;
  }, [requestKey]);
  useEffect(() => {
    cancel();
    controllerRef.current?.abort();
    if (!stableRequest) {
      setAssessment(null);
      setIsAssessing(false);
      return;
    }

    setAssessment(getLocalOkrQualityAssessment(stableRequest));
    if (isReadyForAiQualityAssessment(stableRequest)) assess(stableRequest);
  }, [assess, cancel, stableRequest]);

  useEffect(
    () => () => {
      cancel();
      controllerRef.current?.abort();
    },
    [cancel],
  );

  return { assessment, isAssessing };
};
