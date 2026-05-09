import { describe, it, expect } from "vitest";
import { encodeFrame, decodeFrame } from "../index";

describe("encodeFrame", () => {
  it("produces a 16-byte header followed by JSON payload", () => {
    const frame = encodeFrame(4, 1, { type: "msg", text: "hi" });

    expect(frame.byteLength).toBeGreaterThan(16);
    const view = new DataView(frame.buffer);
    // totalLength (4 bytes)
    expect(view.getUint32(0)).toBe(frame.byteLength);
    // headerLength (2 bytes)
    expect(view.getUint16(4)).toBe(16);
    // version (2 bytes)
    expect(view.getUint16(6)).toBe(1);
    // op (4 bytes)
    expect(view.getUint32(8)).toBe(4);
    // seq (4 bytes)
    expect(view.getUint32(12)).toBe(1);
  });

  it("encodes the body as JSON after the header", () => {
    const frame = encodeFrame(7, 42, { token: "abc", domain: "platform" });
    const payload = new TextDecoder().decode(frame.buffer.slice(16));
    const obj = JSON.parse(payload);
    expect(obj).toEqual({ token: "abc", domain: "platform" });
  });

  it("works with empty body", () => {
    const frame = encodeFrame(2, 0, {});
    expect(frame.byteLength).toBe(16 + 2); // {} is 2 bytes
    const view = new DataView(frame.buffer);
    expect(view.getUint32(8)).toBe(2);
  });
});

describe("decodeFrame", () => {
  it("returns op and body from a valid frame", () => {
    const frame = encodeFrame(4, 3, { msg: "hello" });
    const result = decodeFrame(frame.buffer);

    expect(result.op).toBe(4);
    expect(JSON.parse(result.body)).toEqual({ msg: "hello" });
  });

  it("returns op 0 for an empty-framed heartbeat", () => {
    const buf = new ArrayBuffer(16);
    const view = new DataView(buf);
    view.setUint32(0, 16);
    view.setUint16(4, 16);
    view.setUint16(6, 1);

    const result = decodeFrame(buf);
    expect(result.op).toBe(0);
    expect(result.body).toBe("");
  });
});

describe("round-trip", () => {
  it("encode then decode preserves data", () => {
    const data = { type: "send_msg", to: 1001, content: "hi there" };
    const encoded = encodeFrame(4, 7, data);
    const decoded = decodeFrame(encoded.buffer);

    expect(decoded.op).toBe(4);
    expect(JSON.parse(decoded.body)).toEqual(data);
  });
});
