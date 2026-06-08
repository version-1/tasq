import {
  ChevronDown,
  Ellipsis,
  LayoutDashboard,
  Plus,
  Settings,
  SquareKanban,
  type LucideIcon,
} from "lucide-react";

const icons = {
  "chevron-down": ChevronDown,
  ellipsis: Ellipsis,
  "layout-dashboard": LayoutDashboard,
  plus: Plus,
  settings: Settings,
  "square-kanban": SquareKanban,
} satisfies Record<string, LucideIcon>;

export type IconProxyName = keyof typeof icons;

type IconProxyProps = {
  name: IconProxyName;
  className?: string;
  size?: number;
  strokeWidth?: number;
};

export function IconProxy({
  name,
  className,
  size = 16,
  strokeWidth = 2,
}: IconProxyProps) {
  const Icon = icons[name];

  return (
    <Icon
      aria-hidden="true"
      className={className}
      focusable="false"
      size={size}
      strokeWidth={strokeWidth}
    />
  );
}
