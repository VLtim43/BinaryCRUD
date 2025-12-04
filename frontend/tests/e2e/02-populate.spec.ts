import { test } from '@playwright/test';
import {
  populateInventory,
  deleteAllFiles,
  verifyItemExists,
  verifyItemNotFound,
  verifyOrderExists,
  verifyOrderNotFound,
  verifyPromotionExists,
  verifyPromotionNotFound,
  testDeleteThenReadFails,
  getSeedItems,
  getSeedOrders,
  getSeedPromotions,
  getSeededItem,
  getSeededOrder,
  getSeededPromotion,
} from '../utils/test-helpers';

// Load seed data counts
const itemCount = getSeedItems().length;
const orderCount = getSeedOrders().length;
const promotionCount = getSeedPromotions().length;

// This test suite assumes 01-setup.spec.ts has already populated the inventory
test.describe('02 - Populate Inventory', () => {

  test.describe('Verify populated items', () => {
    test('first item exists', async ({ page }) => {
      await page.goto('/');
      const item = getSeededItem(0);
      await verifyItemExists(page, item.id, item.name);
    });

    test('middle item exists', async ({ page }) => {
      await page.goto('/');
      const item = getSeededItem(Math.floor(itemCount / 2));
      await verifyItemExists(page, item.id, item.name);
    });

    test('last item exists', async ({ page }) => {
      await page.goto('/');
      const item = getSeededItem(itemCount - 1);
      await verifyItemExists(page, item.id, item.name);
    });
  });

  test.describe('Verify populated orders', () => {
    test('first order exists', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(0);
      await verifyOrderExists(page, order.id, order.customer);
    });

    test('middle order exists', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(Math.floor(orderCount / 2));
      await verifyOrderExists(page, order.id, order.customer);
    });

    test('last order exists', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(orderCount - 1);
      await verifyOrderExists(page, order.id, order.customer);
    });
  });

  test.describe('Verify populated promotions', () => {
    test('first promotion exists', async ({ page }) => {
      await page.goto('/');
      const promo = getSeededPromotion(0);
      await verifyPromotionExists(page, promo.id, promo.name);
    });

    test('middle promotion exists', async ({ page }) => {
      await page.goto('/');
      const promo = getSeededPromotion(Math.floor(promotionCount / 2));
      await verifyPromotionExists(page, promo.id, promo.name);
    });

    test('last promotion exists', async ({ page }) => {
      await page.goto('/');
      const promo = getSeededPromotion(promotionCount - 1);
      await verifyPromotionExists(page, promo.id, promo.name);
    });
  });

  test.describe('Delete and verify read fails', () => {
    // Use index 4 (id=5) for deletion tests to avoid first/last items
    const deleteIndex = 4;

    test('delete item and verify read fails', async ({ page }) => {
      await page.goto('/');
      const item = getSeededItem(deleteIndex);
      await testDeleteThenReadFails(page, 'item', item.id);
    });

    test('delete order and verify read fails', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(Math.min(deleteIndex, orderCount - 1));
      await testDeleteThenReadFails(page, 'order', order.id);
    });

    test('delete promotion and verify read fails', async ({ page }) => {
      await page.goto('/');
      const promo = getSeededPromotion(Math.min(deleteIndex, promotionCount - 1));
      await testDeleteThenReadFails(page, 'promotion', promo.id);
    });
  });

  test.describe('Verify other records still exist after deletions', () => {
    test('first item still exists after deletion', async ({ page }) => {
      await page.goto('/');
      const item = getSeededItem(0);
      await verifyItemExists(page, item.id, item.name);
    });

    test('first order still exists after deletion', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(0);
      await verifyOrderExists(page, order.id, order.customer);
    });

    test('first promotion still exists after deletion', async ({ page }) => {
      await page.goto('/');
      const promo = getSeededPromotion(0);
      await verifyPromotionExists(page, promo.id, promo.name);
    });
  });

  test.describe('Cleanup - delete all remaining records', () => {
    test('delete all files', async ({ page }) => {
      await page.goto('/');
      await deleteAllFiles(page);
    });

    test('verify first item no longer exists', async ({ page }) => {
      await page.goto('/');
      const item = getSeededItem(0);
      await verifyItemNotFound(page, item.id);
    });

    test('verify first order no longer exists', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(0);
      await verifyOrderNotFound(page, order.id);
    });

    test('verify first promotion no longer exists', async ({ page }) => {
      await page.goto('/');
      const promo = getSeededPromotion(0);
      await verifyPromotionNotFound(page, promo.id);
    });
  });
});
