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
  placement?: "bottom-end" | "bottom-start";
  size?: "default" | "wide";
  trigger: (props: ContextMenuTriggerProps) => ReactNode;
};

export function ContextMenu({
  boundaryRef,
  children,
  id,
  isOpen,
  label,
  onOpenChange,
  placement = "bottom-end",
  size = "default",
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
        <div
          aria-label={label}
          className={[
            styles.menu,
            styles[placement],
            size === "wide" ? styles.wideMenu : "",
          ].filter(Boolean).join(" ")}
          id={id}
          role="menu"
        >
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
  accessory,
  children,
  disabled = false,
  icon,
  label,
  onSelect,
  selected = false,
  title,
  variant = "default",
}: {
  accessory?: ReactNode;
  children: ReactNode;
  disabled?: boolean;
  icon?: ReactNode;
  label?: string;
  onSelect?: () => void;
  selected?: boolean;
  title?: string;
  variant?: "default" | "danger";
}) {
  return (
    <button
      aria-label={label}
      className={[
        styles.item,
        icon || accessory ? styles.richItem : "",
        variant === "danger" ? styles.dangerItem : "",
        selected ? styles.selectedItem : "",
      ]
        .filter(Boolean)
        .join(" ")}
      data-selected={selected || undefined}
      disabled={disabled}
      role="menuitem"
      title={title}
      type="button"
      onClick={onSelect}
    >
      {icon ? <span className={styles.itemIcon}>{icon}</span> : null}
      <span className={styles.itemContent}>{children}</span>
      {accessory ? <span className={styles.itemAccessory}>{accessory}</span> : null}
    </button>
  );
}

export function ContextMenuSeparator() {
  return <div className={styles.separator} role="separator" />;
}

export function ContextMenuHelp({ children }: { children: ReactNode }) {
  return <p className={styles.help}>{children}</p>;
}
