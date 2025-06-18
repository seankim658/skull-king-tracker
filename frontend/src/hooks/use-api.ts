// Used for fetching data

import { useState, useCallback } from "react";
import { toast } from "sonner";
import { errorExtract } from "@/lib/utils";
import type { ApiResponse } from "@/lib/api/types";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type ApiFunction<T, TArgs extends any[]> = (
  ...args: TArgs
) => Promise<ApiResponse<T>>;

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useApi<T, TArgs extends any[]>(
  apiFunc: ApiFunction<T, TArgs>,
  options?: { onSuccess?: (data: T) => void; onError?: (error: Error) => void },
) {
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const request = useCallback(
    async (...args: TArgs) => {
      setIsLoading(true);
      setError(null);
      try {
        const response = await apiFunc(...args);
        if (response.success && response.data !== undefined) {
          setData(response.data);
          options?.onSuccess?.(response.data);
        } else {
          const errorMessage =
            response.message || "An unexpected error occurred";
          setError(errorMessage);
          toast.error(errorMessage);
          options?.onError?.(new Error(errorMessage));
        }
      } catch (e) {
        const errorMessage = errorExtract(e, "An API error occurred");
        setError(errorMessage);
        toast.error(errorMessage);
        options?.onError?.(e as Error);
      } finally {
        setIsLoading(false);
      }
    },
    [apiFunc, options],
  );

  return { data, isLoading, error, request, setData };
}
