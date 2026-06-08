export type ToastType = "error" | "success";

export type Toast = {
  readonly id: string;
  readonly message: string;
  readonly type: ToastType;
};

export type ToastInput = {
  readonly message: string;
  readonly type: ToastType;
};

export type ToastListener = (toasts: readonly Toast[]) => void;

const toastTimeoutMs: Record<ToastType, number> = {
  error: 5_000,
  success: 3_000,
};

const listeners = new Set<ToastListener>();
const timers = new Map<string, ReturnType<typeof setTimeout>>();

let nextToastID = 1;
let toasts: Toast[] = [];

export function addToast(input: ToastInput): Toast {
  const toast: Toast = {
    id: createToastID(),
    message: input.message,
    type: input.type,
  };

  toasts = [...toasts, toast];
  timers.set(
    toast.id,
    setTimeout(() => {
      removeToast(toast.id);
    }, toastTimeoutMs[toast.type]),
  );
  notifyListeners();

  return toast;
}

export function subscribe(listener: ToastListener): () => void {
  listeners.add(listener);
  listener(getToastsSnapshot());

  return () => {
    listeners.delete(listener);
  };
}

function removeToast(id: string): void {
  const timer = timers.get(id);
  if (timer) {
    clearTimeout(timer);
    timers.delete(id);
  }

  const nextToasts = toasts.filter((toast) => toast.id !== id);
  if (nextToasts.length === toasts.length) {
    return;
  }

  toasts = nextToasts;
  notifyListeners();
}

function createToastID(): string {
  const id = `toast-${nextToastID}`;
  nextToastID += 1;
  return id;
}

function getToastsSnapshot(): readonly Toast[] {
  return [...toasts];
}

function notifyListeners(): void {
  const snapshot = getToastsSnapshot();
  listeners.forEach((listener) => listener(snapshot));
}
