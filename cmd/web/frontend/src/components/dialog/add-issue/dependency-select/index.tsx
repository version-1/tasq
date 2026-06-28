import { useEffect, useRef, useState } from "react";
import { IconProxy } from "@/components/ui/icon-proxy";
import type { IssueSummary } from "@/lib/types";
import styles from "../index.module.css";

export function DependencySelect({
  emptyLabel,
  getStatusLabel,
  label,
  options,
  placeholder,
  resetKey,
  selectedCountLabel,
  selectedIDs,
  onChange,
}: {
  emptyLabel: string;
  getStatusLabel: (issue: IssueSummary) => string;
  label: string;
  options: IssueSummary[];
  placeholder: string;
  resetKey: number | null;
  selectedCountLabel: (count: number) => string;
  selectedIDs: number[];
  onChange: (selectedIDs: number[]) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const selectedOptions = options.filter((issue) => selectedIDs.includes(issue.id));
  const summary = selectedOptions.length === 0 ? placeholder : selectedCountLabel(selectedOptions.length);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    function handlePointerDown(event: PointerEvent) {
      if (!dropdownRef.current?.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setIsOpen(false);
      }
    }

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen]);

  useEffect(() => {
    setIsOpen(false);
  }, [resetKey]);

  return (
    <fieldset className={styles.dependencyField}>
      <legend>{label}</legend>
      <div ref={dropdownRef} className={styles.dependencyDropdown}>
        <button
          type="button"
          aria-controls="add-issue-dependency-options"
          aria-expanded={isOpen}
          className={styles.dependencyTrigger}
          onClick={() => setIsOpen((current) => !current)}
        >
          <span className={styles.dependencySummary}>
            {selectedOptions.length === 0 ? (
              <span className={styles.emptyValue}>{summary}</span>
            ) : (
              selectedOptions.map((issue) => (
                <span key={issue.id} className={styles.dependencyChip}>
                  #{issue.id} {issue.title}
                </span>
              ))
            )}
          </span>
          <IconProxy className={styles.dependencyTriggerIcon} name="chevron-down" size={18} strokeWidth={2.3} />
        </button>
        {isOpen ? (
          <div
            id="add-issue-dependency-options"
            className={styles.dependencyPopover}
            role="group"
            aria-label={label}
          >
            {options.length > 0 ? (
              <div className={styles.dependencyOptions}>
                {options.map((issue) => (
                  <label key={issue.id} className={styles.dependencyOption}>
                    <input
                      type="checkbox"
                      checked={selectedIDs.includes(issue.id)}
                      onChange={() => onChange(toggleDependencyID(selectedIDs, issue.id))}
                    />
                    <span>
                      <strong>#{issue.id}</strong> {issue.title}
                      <small>{getStatusLabel(issue)}</small>
                    </span>
                  </label>
                ))}
              </div>
            ) : (
              <p className={styles.dependencyEmpty}>{emptyLabel}</p>
            )}
          </div>
        ) : null}
      </div>
    </fieldset>
  );
}

function toggleDependencyID(currentIDs: number[], issueID: number): number[] {
  if (currentIDs.includes(issueID)) {
    return currentIDs.filter((currentID) => currentID !== issueID);
  }
  return [...currentIDs, issueID].sort((left, right) => left - right);
}
