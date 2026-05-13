# @game/api

Protocol layer for the VDA + IM frontend. Contains the IM WebSocket binary frame codec and HTTP API clients.

## IM Protocol

### Binary Frame Format

```
Byte 0-3   : totalLength (uint32, big-endian)
Byte 4-5   : headerLength (uint16, always 16)
Byte 6-7   : version (uint16, currently 1)
Byte 8-11  : op (uint32, operation code)
Byte 12-15 : seq (uint32, sequence number)
Byte 16+   : JSON payload (UTF-8)
```

### Opcodes

| Op | Constant | Description |
|----|----------|-------------|
| 2 | OP_HEARTBEAT | Client heartbeat |
| 3 | OP_HEARTBEAT_REPLY | Server heartbeat reply |
| 4 | OP_SEND | Message/push (bidirectional) |
| 7 | OP_AUTH | Client authentication |
| 8 | OP_AUTH_REPLY | Server auth response |

### Usage

```ts
import { encodeFrame, decodeFrame } from "@game/api";

// Encode a frame
const frame = encodeFrame(4, 1, { type: "msg", text: "hello" });
ws.send(frame);

// Decode a received frame
ws.onmessage = (e) => {
  const packet = decodeFrame(e.data); // { op: number, body: string }
};
```

### Auth Request

```ts
type IMAuthRequest = {
  token: string;
  domain: "platform" | "tenant";
  scope: {
    tenant_id: string;
    project_id: string;
    environment: string;
  };
};
```

## HTTP API Clients

```ts
import {
  HttpClient,
  createUserAPI,
  createFriendAPI,
  createCommentAPI,
  createHistoryAPI,
} from "@game/api";

const http = new HttpClient("http://localhost:8080", "jwt-token");
const userAPI = createUserAPI(http);
const profile = await userAPI.getProfile(1001);
const comments = await createCommentAPI(http).list({
  objId: 1001,
  objType: 1,
  pageSize: 20,
  sortType: 0,
});
```

### Endpoints

| API | Methods |
|-----|---------|
| UserAPI | `getProfile`, `searchUsers` |
| FriendAPI | `list`, `add`, `remove` |
| CommentAPI | `list`, `get`, `create`, `delete`, `like`, `unlike`, `block`, `unblock`, `pin`, `unpin` |
| HistoryAPI | `list` |
| PlatformAPI | `getDemoToken`, `createTenant`, `getTenant`, `updateTenant`, `deleteTenant`, `myTenants`, `createProject`, `listProjects`, `createEnvironment`, `listEnvironments` |
