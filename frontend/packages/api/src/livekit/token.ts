const encoder = new TextEncoder();

function base64url(buf: ArrayBuffer | Uint8Array): string {
  const bytes = buf instanceof Uint8Array ? buf : new Uint8Array(buf);
  let binary = "";
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

function strToBuf(s: string): ArrayBuffer {
  return encoder.encode(s).buffer as ArrayBuffer;
}

async function hmacSign(key: ArrayBuffer, data: ArrayBuffer): Promise<ArrayBuffer> {
  const cryptoKey = await crypto.subtle.importKey(
    "raw",
    key,
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"],
  );
  return crypto.subtle.sign("HMAC", cryptoKey, data);
}

type VideoGrant = {
  room?: string;
  roomJoin?: boolean;
  canPublish?: boolean;
  canSubscribe?: boolean;
};

type TokenOptions = {
  roomName: string;
  participantId: string;
  apiKey: string;
  apiSecret: string;
  ttl?: number; // seconds, default 3600
};

export async function generateLiveKitToken(opts: TokenOptions): Promise<string> {
  const { roomName, participantId, apiKey, apiSecret, ttl = 3600 } = opts;

  const header = { alg: "HS256", typ: "JWT" };
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    iss: apiKey,
    sub: participantId,
    nbf: now,
    exp: now + ttl,
    video: {
      room: roomName,
      roomJoin: true,
      canPublish: true,
      canSubscribe: true,
    } satisfies VideoGrant,
  };

  const headerB64 = base64url(strToBuf(JSON.stringify(header)));
  const payloadB64 = base64url(strToBuf(JSON.stringify(payload)));
  const signingInput = strToBuf(`${headerB64}.${payloadB64}`);
  const secret = strToBuf(apiSecret);

  const signature = await hmacSign(secret, signingInput);
  const sigB64 = base64url(signature);

  return `${headerB64}.${payloadB64}.${sigB64}`;
}
