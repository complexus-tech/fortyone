import { cn } from "lib";
import { ReactNode } from "react";
import { DropzoneInputProps, DropzoneRootProps } from "react-dropzone";
import { Text } from "../Text/Text";
import { Box } from "../Box/Box";

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
          className="inline-block mb-3 font-medium dark:text-white"
        >
          {label}
          {required && <span className="text-danger">*</span>}
        </label>
      )}
      {children}
    </div>
  );
};

export type RootProps = {
  rootProps: DropzoneRootProps;
  isDragActive: boolean;
  children?: ReactNode;
};
const Root = (props: RootProps) => {
  const { children, isDragActive, rootProps } = props;
  const { className, ...rest } = rootProps;
  return (
    <div
      className={cn(
        "border-[1.5px] border-dashed hover:border-primary/40 transition dark:hover:border-primary/40 border-gray-200 bg-gray-50/50 dark:border-dark-100 rounded-xl dark:bg-dark-200/40 h-24 cursor-pointer flex justify-center items-center p-4",
        {
          "bg-gray-50 dark:bg-dark-200/80 hover:border-primary/40 transition dark:hover:border-primary/40":
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

<Box className="mb-4 mt-3 flex h-24 cursor-pointer items-center justify-center rounded-xl border-[1.5px] border-dashed border-gray-200 bg-gray-50/50 dark:border-dark-100 dark:bg-dark-200/40">
  <Text color="muted">Click or drag files here</Text>
</Box>;

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
