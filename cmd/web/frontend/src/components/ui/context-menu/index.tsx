import { useEffect, useRef, type ReactNode, type RefObject } from "react";
import styles from "./index.module.css";

type ContextMenuTriggerProps = {
  "aria-controls"?: string;
  "aria-expanded": boolean;
  "aria-haspopup": "menu";
  onClick: () => void;
};

type ContextMenuProps = {
  boundaryRef?: RefObject<HTMLElement | null>;
  children: ReactNode;
  id: string;
  isOpen: boolean;
  label: string;
  onOpenChange: (isOpen: boolean) => void;
  trigger: (props: ContextMenuTriggerProps) => ReactNode;
};

export function ContextMenu({
  boundaryRef,
  children,
  id,
  isOpen,
  label,
  onOpenChange,
  trigger,
}: ContextMenuProps) {
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    function closeMenuOnOutsidePointerDown(event: PointerEvent) {
      const boundary = boundaryRef?.current ?? rootRef.current;
      if (!boundary || !(event.target instanceof Node) || boundary.contains(event.target)) {
        return;
      }
      onOpenChange(false);
    }

    document.addEventListener("pointerdown", closeMenuOnOutsidePointerDown);
    return () => {
      document.removeEventListener("pointerdown", closeMenuOnOutsidePointerDown);
    };
  }, [boundaryRef, isOpen, onOpenChange]);

  return (
    <div className={styles.root} ref={rootRef}>
      {trigger({
        "aria-controls": isOpen ? id : undefined,
        "aria-expanded": isOpen,
        "aria-haspopup": "menu",
        onClick: () => onOpenChange(!isOpen),
      })}
      {isOpen ? (
        <div aria-label={label} className={styles.menu} id={id} role="menu">
          {children}
        </div>
      ) : null}
    </div>
  );
}

export function ContextMenuGroupLabel({ children }: { children: ReactNode }) {
  return <div className={styles.groupLabel}>{children}</div>;
}

export function ContextMenuItem({
  children,
  disabled = false,
  label,
  onSelect,
  title,
}: {
  children: ReactNode;
  disabled?: boolean;
  label?: string;
  onSelect?: () => void;
  title?: string;
}) {
  return (
    <button
      aria-label={label}
      className={styles.item}
      disabled={disabled}
      role="menuitem"
      title={title}
      type="button"
      onClick={onSelect}
    >
      {children}
    </button>
  );
}

export function ContextMenuHelp({ children }: { children: ReactNode }) {
  return <p className={styles.help}>{children}</p>;
}
