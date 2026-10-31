import { useRef, useState } from "react";
import type { AutoOKChangeHandler } from "../components/issues-view/types";

export function useAutoOKToggle(issueID: number, onChange?: AutoOKChangeHandler) {
  const [isSaving, setIsSaving] = useState(false);
  const [hasError, setHasError] = useState(false);
  const saving = useRef(false);

  async function toggle(enabled: boolean) {
    if (saving.current || !onChange) return;
    saving.current = true;
    setIsSaving(true);
    setHasError(false);
    try {
      await onChange(issueID, enabled);
    } catch {
      setHasError(true);
    } finally {
      saving.current = false;
      setIsSaving(false);
    }
  }

  return { isSaving, hasError, toggle };
}
