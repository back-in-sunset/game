import { test, expect } from "@playwright/test";

const ROOM_ID = `e2e-room-${Date.now()}`;

/**
 * Navigate to a demo room and wait until IM is connected.
 */
async function goToRoom(page: import("@playwright/test").Page, roomId: string) {
  await page.goto(`/demo/room/${roomId}`);
  await expect(page.getByText("Demo Room")).toBeVisible();
  await expect(page.getByText("IM: 已连接")).toBeVisible({ timeout: 15000 });
}

/**
 * Send a message and wait for it to appear (substring match on text).
 * Returns the full text content of the matching element.
 */
async function sendAndWait(page: import("@playwright/test").Page, text: string) {
  const form = page.locator("form.composer");
  await form.locator('input[name="message"]').fill(text);
  await form.locator('button[type="submit"]').click();
  await expect(page.getByText(text)).toBeVisible({ timeout: 10000 });
}

test.describe("Demo Room IM", () => {
  test("single page: send and see own message, no duplicates", async ({ page }) => {
    await goToRoom(page, ROOM_ID + "-single");

    const uniqueText = `uniq-${Date.now()}`;
    await sendAndWait(page, uniqueText);

    // Wait for any server echo to arrive, then count
    await page.waitForTimeout(1500);
    const count = await page.getByText(uniqueText).count();
    expect(count).toBe(1);
  });

  test("two pages: each can send and see own messages independently", async ({ browser }) => {
    const ctxA = await browser.newContext();
    const ctxB = await browser.newContext();
    const pageA = await ctxA.newPage();
    const pageB = await ctxB.newPage();

    await goToRoom(pageA, ROOM_ID);
    await goToRoom(pageB, ROOM_ID);

    // Each page sends independently
    const textA = `a-${Date.now()}`;
    const textB = `b-${Date.now()}`;
    await sendAndWait(pageA, textA);
    await sendAndWait(pageB, textB);

    // Verify messages persisted (not cleared by history load or re-render)
    await pageA.waitForTimeout(2000);
    await pageB.waitForTimeout(2000);

    await expect(pageA.getByText(textA)).toBeVisible({ timeout: 5000 });
    await expect(pageB.getByText(textB)).toBeVisible({ timeout: 5000 });

    await ctxA.close();
    await ctxB.close();
  });

  test("cross-tab: page B receives page A messages in real-time", async ({ browser }) => {
    const ctxA = await browser.newContext();
    const ctxB = await browser.newContext();
    const pageA = await ctxA.newPage();
    const pageB = await ctxB.newPage();

    await goToRoom(pageA, ROOM_ID + "-cross");
    await goToRoom(pageB, ROOM_ID + "-cross");

    // Page A sends message — it should arrive at Page B via IM push
    const crossText = `cross-${Date.now()}`;
    await sendAndWait(pageA, crossText);

    // Give IM push time to deliver
    await expect(pageB.getByText(crossText)).toBeVisible({ timeout: 20000 });

    await ctxA.close();
    await ctxB.close();
  });

  test("history: new tab sees messages sent before it joined", async ({ browser }) => {
    const ctxA = await browser.newContext();
    const pageA = await ctxA.newPage();
    const historyRoomId = ROOM_ID + "-hist";

    await goToRoom(pageA, historyRoomId);

    // Send messages from Page A
    const msgs = [`h1-${Date.now()}`, `h2-${Date.now()}`];
    await sendAndWait(pageA, msgs[0]);
    await sendAndWait(pageA, msgs[1]);

    // Small delay to ensure server persistence
    await pageA.waitForTimeout(500);

    // Page B joins after — should load history via list_messages
    const ctxB = await browser.newContext();
    const pageB = await ctxB.newPage();
    await goToRoom(pageB, historyRoomId);

    for (const m of msgs) {
      await expect(pageB.getByText(m)).toBeVisible({ timeout: 20000 });
    }

    await ctxA.close();
    await ctxB.close();
  });
});
