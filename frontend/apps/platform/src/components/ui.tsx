import type {
  ButtonHTMLAttributes,
  HTMLAttributes,
  InputHTMLAttributes,
  LabelHTMLAttributes,
  ReactNode,
} from "react";
import { cn } from "../lib/cn";

type CardProps = HTMLAttributes<HTMLElement> & {
  title: string;
  description?: string;
  action?: ReactNode;
};

export function Card({ title, description, action, className, children, ...props }: CardProps) {
  return (
    <section
      className={cn(
        "rounded-3xl border border-slate-200/70 bg-white/80 p-5 shadow-[0_18px_48px_rgba(15,23,42,0.06)] backdrop-blur",
        className,
      )}
      {...props}
    >
      <header className="mb-4 flex items-start justify-between gap-3">
        <div className="space-y-1">
          <h2 className="text-base font-semibold tracking-tight text-slate-950">{title}</h2>
          {description ? <p className="text-sm text-slate-500">{description}</p> : null}
        </div>
        {action ? <div className="shrink-0">{action}</div> : null}
      </header>
      {children}
    </section>
  );
}

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "ghost" | "destructive";
  size?: "sm" | "md";
};

export function Button({
  variant = "secondary",
  size = "md",
  className,
  type = "button",
  ...props
}: ButtonProps) {
  const variants: Record<NonNullable<ButtonProps["variant"]>, string> = {
    primary: "border-transparent bg-teal-600 text-white hover:bg-teal-700",
    secondary: "border-slate-200 bg-white text-slate-900 hover:bg-slate-50",
    ghost: "border-transparent bg-transparent text-slate-700 hover:bg-slate-100",
    destructive: "border-transparent bg-rose-600 text-white hover:bg-rose-700",
  };
  const sizes: Record<NonNullable<ButtonProps["size"]>, string> = {
    sm: "h-8 px-3 text-xs",
    md: "h-10 px-4 text-sm",
  };

  return (
    <button
      type={type}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-xl border font-medium transition-colors disabled:pointer-events-none disabled:opacity-50",
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    />
  );
}

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm text-slate-900 outline-none transition-shadow placeholder:text-slate-400 focus:border-teal-500 focus:ring-4 focus:ring-teal-500/10",
        className,
      )}
      {...props}
    />
  );
}

type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  tone?: "neutral" | "success" | "warning" | "danger" | "info";
};

export function Badge({ tone = "neutral", className, ...props }: BadgeProps) {
  const tones: Record<NonNullable<BadgeProps["tone"]>, string> = {
    neutral: "bg-slate-100 text-slate-700",
    success: "bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200",
    warning: "bg-amber-50 text-amber-800 ring-1 ring-amber-200",
    danger: "bg-rose-50 text-rose-700 ring-1 ring-rose-200",
    info: "bg-cyan-50 text-cyan-700 ring-1 ring-cyan-200",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium",
        tones[tone],
        className,
      )}
      {...props}
    />
  );
}

export function Label({ className, ...props }: LabelHTMLAttributes<HTMLLabelElement>) {
  return <label className={cn("text-sm font-medium text-slate-700", className)} {...props} />;
}

export function Separator({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("h-px bg-slate-200/80", className)} {...props} />;
}

export function EmptyState({
  title,
  description,
}: {
  title: string;
  description?: string;
}) {
  return (
    <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50/80 p-4 text-sm text-slate-500">
      <div className="font-medium text-slate-700">{title}</div>
      {description ? <div className="mt-1 leading-6">{description}</div> : null}
    </div>
  );
}
