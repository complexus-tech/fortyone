import { ErrorIcon, InfoIcon, SuccessIcon, WarningIcon } from "icons";

export const toasterIcons = {
  success: <SuccessIcon className="h-6" />,
  info: <InfoIcon className="h-6" />,
  warning: <WarningIcon className="text-warning dark:text-warning h-6" />,
  error: <ErrorIcon className="text-danger dark:text-danger h-6" />,
};
