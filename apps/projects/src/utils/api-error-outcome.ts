export type ApiErrorOutcomeReport = {
  certainty: "definite" | "uncertain";
  status?: number;
};

type ApiErrorOutcomeReporter = (report: ApiErrorOutcomeReport) => void;

const API_ERROR_OUTCOME_REPORTER_KEY = Symbol.for(
  "fortyone.api-error-outcome-reporter",
);

type GlobalWithApiErrorOutcomeReporter = typeof globalThis & {
  [API_ERROR_OUTCOME_REPORTER_KEY]?: ApiErrorOutcomeReporter;
};

const getGlobalReporterState = () =>
  globalThis as GlobalWithApiErrorOutcomeReporter;

/**
 * Installs the server-side observer used by Maya's approved-mutation boundary.
 * Browser and ordinary server-action callers never install an observer, so
 * reporting has no effect on their response shape or control flow.
 */
export const installApiErrorOutcomeReporter = (
  reporter: ApiErrorOutcomeReporter,
) => {
  getGlobalReporterState()[API_ERROR_OUTCOME_REPORTER_KEY] = reporter;
};

export const reportApiErrorOutcome = (report: ApiErrorOutcomeReport) => {
  getGlobalReporterState()[API_ERROR_OUTCOME_REPORTER_KEY]?.(report);
};
