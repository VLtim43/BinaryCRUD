import { test, expect } from '@playwright/test';

test.describe('01 - Initial Setup', () => {
  test('app loads successfully', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('body')).toBeVisible();
    await expect(page.locator('#App')).toBeVisible();
  });

  test('all tabs are visible', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('button.tab:has-text("Item")')).toBeVisible();
    await expect(page.locator('button.tab:has-text("Order")')).toBeVisible();
    await expect(page.locator('button.tab:has-text("Promotion")')).toBeVisible();
    await expect(page.locator('button.tab:has-text("Debug")')).toBeVisible();
  });

  test('delete all files (clean state)', async ({ page }) => {
    await page.goto('/');
    await page.click('button.tab:has-text("Debug")');
    await page.click('button:has-text("Delete All Files")');
    // Wait for deletion to complete
    await page.waitForTimeout(2000);
  });

  test('populate inventory from seed data', async ({ page }) => {
    await page.goto('/');
    await page.click('button.tab:has-text("Debug")');
    await page.click('button:has-text("Populate Inventory")');
    await expect(page.locator('text=Inventory populated successfully')).toBeVisible({ timeout: 15000 });
  });
});
