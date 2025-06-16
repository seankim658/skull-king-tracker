import type { ReactNode } from "react";
import { useState, useCallback, useMemo } from "react";
import {
  ConfirmContext,
  type ConfirmOptions,
} from "@/contexts/confirm-context";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

interface ConfirmProviderProps {
  children: ReactNode;
}

export function ConfirmProvider({ children }: ConfirmProviderProps) {
  const [options, setOptions] = useState<ConfirmOptions | null>(null);
  const [resolve, setResolve] = useState<((value: boolean) => void) | null>(
    null,
  );

  const confirm = useCallback((newOptions: ConfirmOptions) => {
    return new Promise<boolean>((res) => {
      setOptions(newOptions);
      setResolve(() => res);
    });
  }, []);

  const handleClose = useCallback(() => {
    if (resolve) {
      resolve(false);
    }
    setOptions(null);
    setResolve(null);
  }, [resolve]);

  const handleConfirm = useCallback(() => {
    if (resolve) {
      resolve(true);
    }
    setOptions(null);
    setResolve(null);
  }, [resolve]);

  const contextValue = useMemo(() => confirm, [confirm]);

  return (
    <ConfirmContext.Provider value={contextValue}>
      {children}
      {options && (
        <ConfirmDialog
          isOpen={!!options}
          onClose={handleClose}
          onConfirm={handleConfirm}
          {...options}
        />
      )}
    </ConfirmContext.Provider>
  );
}
