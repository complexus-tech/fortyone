type WorkPlanResponse<T> = {
  data?: T | null;
  error?: { message?: string | null } | null;
};

type WorkPlanResponseResult<T> = { data: T } | { error: string };

export const resolveMayaWorkPlanResponse = <T>(
  response: WorkPlanResponse<T>,
  missingDataMessage: string,
): WorkPlanResponseResult<T> => {
  const apiError = response.error?.message?.trim();
  if (apiError) return { error: apiError };
  if (!response.data) return { error: missingDataMessage };
  return { data: response.data };
};
