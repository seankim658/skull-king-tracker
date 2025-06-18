// Used for performing actions (POST, PUT, DELETE)

import { useState, useCallback } from "react";
import { toast } from "sonner";
import { errorExtract } from "@/lib/utils";
import type { ApiResponse } from "@/lib/api/types";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type ApiSubmitFunction<T, TArgs extends any[]> = (
  ...args: TArgs
) => Promise<ApiResponse<T>>;

// eslint-disable-next-line @typescript-eslint/no-explicit-any
interface UseSubmitOptions<T, TArgs extends any[]> {
  actionVerb: string;
  onSuccess?: (data?: T | undefined, ...args: TArgs) => void;
  onError?: (error: Error) => void;
  successMessage?: string;
  errorMessage?: string;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useSubmit<T, TArgs extends any[]>(
  apiFunc: ApiSubmitFunction<T, TArgs>,
  options?: UseSubmitOptions<T, TArgs>,
) {
  const [isLoading, setIsLoading] = useState(false);

  const submit = useCallback(
    async (...args: TArgs) => {
      setIsLoading(true);
      const progressMsg = options
        ? `${options.actionVerb}...`
        : "Processing...";
      const toastId = toast.loading(progressMsg);
      try {
        const response = await apiFunc(...args);
        if (response.success) {
          toast.success(
            options?.successMessage || response.message || "Success!",
            { id: toastId },
          );
          options?.onSuccess?.(response.data, ...args);
        } else {
          const errMsg =
            options?.errorMessage || response.message || "An error occurred";
          toast.error(errMsg, { id: toastId });
          options?.onError?.(new Error(errMsg));
        }
      } catch (e) {
        const errMsg = errorExtract(
          e,
          options?.errorMessage || "An error occurred",
        );
        toast.error(errMsg, { id: toastId });
        options?.onError?.(e as Error);
      } finally {
        setIsLoading(false);
      }
    },
    [apiFunc, options],
  );

  return { submit, isLoading };
}
