import React from "react";
import { TextInput } from "react-native";

export const Title = () => {
  return (
    <TextInput
      placeholder="Enter title"
      className="text-4xl font-medium dark:placeholder:text-gray-300 placeholder:text-gray dark:text-white"
      autoFocus
      multiline
    />
  );
};
