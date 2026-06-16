import {
  ArrowDown,
  ArrowRight,
  ArrowUp,
  ArrowUpDown,
  ChevronDown,
  Ellipsis,
  Filter,
  Folder,
  History,
  LayoutGrid,
  LayoutDashboard,
  MessageSquare,
  Plus,
  Settings,
  SquareKanban,
  type LucideIcon,
} from "lucide-react";

const icons = {
  "arrow-down": ArrowDown,
  "arrow-right": ArrowRight,
  "arrow-up": ArrowUp,
  "arrow-up-down": ArrowUpDown,
  "chevron-down": ChevronDown,
  ellipsis: Ellipsis,
  filter: Filter,
  folder: Folder,
  history: History,
  "layout-grid": LayoutGrid,
  "layout-dashboard": LayoutDashboard,
  "message-square": MessageSquare,
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
