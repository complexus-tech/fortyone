import { ErrorIcon, InfoIcon, SuccessIcon, WarningIcon } from "icons";

export const toasterIcons = {
  success: <SuccessIcon className="h-6 text-success dark:text-success" />,
  info: <InfoIcon className="h-6" />,
  warning: <WarningIcon className="h-6 text-warning dark:text-warning" />,
  error: <ErrorIcon className="h-6 text-danger dark:text-danger" />,
};
