import type { ButtonHTMLAttributes, ReactNode } from "react";
import styles from "./index.module.css";

export type ButtonVariant = "primary" | "positive" | "secondary" | "tertiary";
export type ButtonSize = "default" | "compact";

type Props = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "type"> & {
  children?: ReactNode;
  size?: ButtonSize;
  type?: "button" | "submit" | "reset";
  variant?: ButtonVariant;
};

export function Button({
  children,
  className,
  size = "default",
  type = "button",
  variant = "primary",
  ...buttonProps
}: Props) {
  return (
    <button
      {...buttonProps}
      className={[styles.button, styles[variant], styles[size], className].filter(Boolean).join(" ")}
      type={type}
    >
      {children}
    </button>
  );
}
