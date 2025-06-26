import { useEffect } from "react";
import { useApi } from "./use-api";
import type { ApiResponse } from "@/lib/api/types";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type ApiFunction<T, TArgs extends any[]> = (
  ...args: TArgs
) => Promise<ApiResponse<T>>;

/**
 * Custom hook that wraps useApi to automatically fetch data when the component mounts or when
 * the provided arguments change.
 *
 * @template T - The expected data type from the API response
 * @template TArgs - The type fo the argument for the API function
 * @param {ApiFunction<T, TArgs>} - The API function to execute
 * @param {...TArgs} - The arguments to pass to the API function
 * @returns The same return values as the useApi hook
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function useFetchOnMount<T, TArgs extends any[]>(
  apiFunc: ApiFunction<T, TArgs>,
  ...args: TArgs
) {
  const { data, isLoading, error, request, setData } = useApi(apiFunc);

  useEffect(() => {
    request(...args);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [request, ...args]);

  return { data, isLoading, error, refetch: request, setData };
}
