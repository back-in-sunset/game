import { test, expect } from "@playwright/test";

test("home page loads and shows dashboard", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByText("VDA + IM Studio")).toBeVisible();
  await expect(page.getByText("会话列表")).toBeVisible();
});

test("navigates to settings page", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "设置" }).click();
  await expect(page.locator("h1")).toContainText("设置");
});

test("navigates to voice page", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "语音通话", exact: true }).click();
  await expect(page.locator("h1")).toContainText("语音通话");
});

test("navigates to friends page", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "好友" }).click();
  await expect(page.getByText("好友列表")).toBeVisible();
});

test("login page renders", async ({ page }) => {
  await page.goto("/login");
  await expect(page.getByText("登录以使用 IM 和语音通话")).toBeVisible();
  await expect(page.getByPlaceholder("http://localhost:8080")).toBeVisible();
});
