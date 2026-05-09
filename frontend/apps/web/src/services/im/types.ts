export type IMListener = (packet: { op: number; body: string }) => void;
export type IMStatusListener = (status: import("@game/api").IMClientStatus) => void;
