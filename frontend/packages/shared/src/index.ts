export type Message = {
  id: string;
  text: string;
  mine?: boolean;
  time: string;
};

export type Conversation = {
  id: string;
  userId: number;
  title: string;
  preview: string;
  subtitle: string;
  online: boolean;
  statusText: string;
  messages: Message[];
  unread: number;
};

export const conversations: Conversation[] = [
  {
    id: "me",
    userId: 1,
    title: "自己 (多设备)",
    preview: "当前设备的消息会同步到其他设备",
    subtitle: "跨设备消息同步",
    online: true,
    statusText: "IM 已连接，消息可在多个标签页同步",
    messages: [
      { id: "s1", text: "打开两个浏览器标签页，发消息试试看", time: "17:00" },
    ],
    unread: 0,
  },
  {
    id: "alice",
    userId: 2002,
    title: "Alice",
    preview: "今晚 8 点开会吗？",
    subtitle: "产品讨论",
    online: true,
    statusText: "设备在线，支持语音通话",
    messages: [
      { id: "m1", text: "今晚 8 点开会吗？", time: "19:20" },
      { id: "m2", text: "可以，我这边先准备通话房间。", mine: true, time: "19:21" },
      { id: "m3", text: "好，我先上线。", time: "19:22" },
    ],
    unread: 2,
  },
  {
    id: "support",
    userId: 3001,
    title: "Support",
    preview: "语音通话状态同步完成",
    subtitle: "系统通知",
    online: true,
    statusText: "VDA 已准备好接收通话事件",
    messages: [
      { id: "m4", text: "语音通话状态同步完成。", time: "09:12" },
      { id: "m5", text: "收到，准备接入前端面板。", mine: true, time: "09:13" },
    ],
    unread: 0,
  },
];

export const callHighlights = [
  { label: "在线会话", value: "128", help: "IM 连接" },
  { label: "通话房间", value: "24", help: "VDA 房间" },
  { label: "消息延迟", value: "48ms", help: "实时感知" },
];
