import { expect, test } from "@playwright/test";

/**
 * Playwright smoke aligned with DESIGN §13.
 * Requires a running Bedrock (make dev-backend or embed binary) and web UI at E2E_BASE_URL.
 * Default: http://127.0.0.1:8080 (embed) or Vite :5173 with proxy.
 */
const user = "admin";
const pass = "admin123";

test.describe("GA smoke", () => {
  test("login → menu → navigate build runs", async ({ page }) => {
    await page.goto("/login");
    await expect(page.getByRole("heading", { name: "Bedrock" })).toBeVisible();

    await page.getByLabel("用户名").fill(user);
    await page.getByLabel("密码").fill(pass);
    await page.getByRole("button", { name: "登录" }).click();

    await expect(page).not.toHaveURL(/\/login/);

    const nav = page.locator(".app-nav");
    await expect(nav).toBeVisible();

    await page.goto("/cicd/build-runs");
    await expect(page).not.toHaveURL(/\/login/);
  });
});
