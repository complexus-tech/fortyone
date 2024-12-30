import type { Icon } from "./types";

export const DeleteIcon = (props: Icon) => {
  return (
    <svg
      {...props}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="m20 9 -1.995 11.3463C17.8369 21.3026 17.0062 22 16.0353 22H7.96474c-0.97095 0 -1.80164 -0.6974 -1.96978 -1.6537L4 9"
        fill="currentColor"
        strokeWidth={1}
      />
      <path
        d="m20 9 -1.995 11.3463C17.8369 21.3026 17.0062 22 16.0353 22H7.96474c-0.97095 0 -1.80164 -0.6974 -1.96978 -1.6537L4 9h16Z"
        stroke="currentColor"
        strokeWidth={1}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M21 6h-5.625M3 6h5.625m0 0V4c0 -1.10457 0.89543 -2 2 -2h2.75c1.1046 0 2 0.89543 2 2v2m-6.75 0h6.75"
        stroke="currentColor"
        strokeWidth={1}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
