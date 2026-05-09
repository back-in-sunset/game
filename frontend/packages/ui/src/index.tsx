import type { FormEvent, ReactNode } from "react";

export function Card({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle?: string;
  children: ReactNode;
}) {
  return (
    <section className="card">
      <header className="cardHeader">
        <h3>{title}</h3>
        {subtitle ? <p>{subtitle}</p> : null}
      </header>
      {children}
    </section>
  );
}

export function Badge({
  tone = "neutral",
  children,
}: {
  tone?: "neutral" | "success" | "warning";
  children: ReactNode;
}) {
  return <span className={`badge ${tone}`}>{children}</span>;
}

export function Button({
  variant = "secondary",
  onClick,
  children,
}: {
  variant?: "primary" | "secondary" | "ghost";
  onClick?: () => void;
  children: ReactNode;
}) {
  return (
    <button className={`button ${variant}`} onClick={onClick} type="button">
      {children}
    </button>
  );
}

export function Avatar({ name }: { name: string }) {
  return <div className="avatar">{name.slice(0, 1).toUpperCase()}</div>;
}

export function Divider() {
  return <hr style={{ border: 0, borderTop: "1px solid rgba(15, 23, 42, 0.08)" }} />;
}

export function Composer({
  placeholder,
  onSend,
}: {
  placeholder: string;
  onSend?: (text: string) => void;
}) {
  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const input = form.elements.namedItem("message") as HTMLInputElement | null;
    const text = input?.value.trim();
    if (!text) return;
    onSend?.(text);
    form.reset();
  }

  return (
    <form className="composer" onSubmit={handleSubmit}>
      <input name="message" placeholder={placeholder} />
      <button type="submit">发送</button>
    </form>
  );
}

export function ConversationRow({
  conversation,
  active,
  onClick,
}: {
  conversation: { id: string; title: string; preview: string; online: boolean; unread?: number };
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button className={`conversationRow ${active ? "active" : ""}`} onClick={onClick}>
      <Avatar name={conversation.title} />
      <div className="conversationMeta">
        <strong>{conversation.title}</strong>
        <span>{conversation.preview}</span>
      </div>
      {conversation.unread ? <span className="unreadBadge">{conversation.unread}</span> : null}
      <Badge tone={conversation.online ? "success" : "neutral"}>
        {conversation.online ? "在线" : "离线"}
      </Badge>
    </button>
  );
}

export function MessageBubble({ message }: { message: { id: string; text: string; mine?: boolean; time: string } }) {
  return (
    <div className={`bubble ${message.mine ? "mine" : ""}`}>
      <p>{message.text}</p>
      <span>{message.time}</span>
    </div>
  );
}

export function StatusRow({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone: "success" | "warning";
}) {
  return (
    <div className="statusRow">
      <span>{label}</span>
      <Badge tone={tone}>{value}</Badge>
    </div>
  );
}

