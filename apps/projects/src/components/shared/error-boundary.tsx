"use client";

import { Component, type ReactNode, type ErrorInfo } from "react";
import { Button } from "ui";
import { WarningIcon } from "icons";
import posthog from "posthog-js";

type ErrorBoundaryProps = {
  children: ReactNode;
  fallback?: ReactNode;
};

type ErrorBoundaryState = {
  hasError: boolean;
  error: Error | null;
};

class ErrorBoundaryImpl extends Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Send to PostHog
    posthog.captureException(error, {
      errorInfo,
    });
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div>
          <WarningIcon aria-hidden="true" className="text-danger" />
          <div className="text-center">
            <h3 className="text-lg font-medium">Something went wrong</h3>
            <p className="mt-1 text-sm">
              {this.state.error?.message || "An unexpected error occurred"}
            </p>
          </div>
          <Button
            className="mt-4"
            onClick={() => {
              this.setState({ hasError: false, error: null });
            }}
            variant="outline"
          >
            Try again
          </Button>
        </div>
      );
    }

    return this.props.children;
  }
}

// Functional component wrapper for better DX
export const ErrorBoundary = ({ children, fallback }: ErrorBoundaryProps) => {
  return <ErrorBoundaryImpl fallback={fallback}>{children}</ErrorBoundaryImpl>;
};
