import { cn } from "lib";
import { ReactNode } from "react";
import { DropzoneInputProps, DropzoneRootProps } from "react-dropzone";
import { Text } from "./text";
import { Box } from "./box";

export const DropZone = ({
  label,
  children,
  required,
}: {
  label?: string;
  required?: boolean;
  children?: ReactNode;
}) => {
  return (
    <div>
      {label && (
        <label
          htmlFor=""
          className="mb-3 inline-block font-medium dark:text-white"
        >
          {label}
          {required && <span className="text-danger">*</span>}
        </label>
      )}
      {children}
    </div>
  );
};

export type RootProps = DropzoneRootProps & {
  rootProps: DropzoneRootProps;
  isDragActive: boolean;
  children?: ReactNode;
};
const Root = (props: RootProps) => {
  const { children, isDragActive, rootProps, className } = props;
  const { ...rest } = rootProps;
  return (
    <div
      className={cn(
        "flex h-24 cursor-pointer items-center justify-center rounded-2xl border border-dashed border-gray-300 bg-surface p-4 transition hover:border-primary/80 dark:border-dark-100 dark:hover:border-primary/40",
        {
          "bg-gray-50 transition hover:border-primary/40 dark:bg-dark-200/80 dark:hover:border-primary/40":
            isDragActive,
        },
        className
      )}
      {...rest}
    >
      {children}
    </div>
  );
};

type InputProps = {
  inputProps: DropzoneInputProps;
};
const Input = (props: InputProps) => {
  const { inputProps } = props;
  return (
    <input
      {...inputProps}
      style={{
        display: "inline",
        position: "absolute",
        opacity: 0,
      }}
      name="file"
    />
  );
};

const Body = ({
  isDragActive,
  message = "Drag 'n' drop some files here, or click to select files.",
  children,
}: {
  isDragActive: boolean;
  message?: string;
  children?: ReactNode;
}) => {
  return (
    <>
      {children ? (
        children
      ) : (
        <>
          {isDragActive ? (
            <Text color="muted">Drop the files here ...</Text>
          ) : (
            <Text color="muted">{message}</Text>
          )}
        </>
      )}
    </>
  );
};

DropZone.Root = Root;
DropZone.Input = Input;
DropZone.Body = Body;
