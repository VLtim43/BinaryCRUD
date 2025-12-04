import { test, expect } from '@playwright/test';
import {
  populateInventory,
  deleteAllFiles,
  goToCompressTab,
  compressFile,
  decompressFile,
  deleteCompressedFile,
  getCompressedFilesCount,
  getBinFilesCount,
  verifyCompressedFileExists,
  verifyCompressedFileNotExists,
  verifyItemExists,
  verifyOrderExists,
  verifyPromotionExists,
  getSeededItem,
  getSeededOrder,
  getSeededPromotion,
} from '../utils/test-helpers';

test.describe('03 - Compression', () => {

  test.describe('Setup', () => {
    test('delete all files for clean state', async ({ page }) => {
      await page.goto('/');
      await deleteAllFiles(page);
    });

    test('populate inventory', async ({ page }) => {
      await page.goto('/');
      await populateInventory(page);
    });

    test('verify bin files exist', async ({ page }) => {
      await page.goto('/');
      const binCount = await getBinFilesCount(page);
      expect(binCount).toBeGreaterThan(0);
    });

    test('no compressed files initially', async ({ page }) => {
      await page.goto('/');
      const compressedCount = await getCompressedFilesCount(page);
      expect(compressedCount).toBe(0);
    });
  });

  test.describe('Huffman Compression', () => {
    test('compress items.bin with Huffman', async ({ page }) => {
      await page.goto('/');
      await compressFile(page, 'items.bin', 'huffman');
    });

    test('verify items.bin.huffman.compressed exists', async ({ page }) => {
      await page.goto('/');
      await verifyCompressedFileExists(page, 'items.bin.huffman.compressed');
    });

    test('compressed file shows correct algorithm', async ({ page }) => {
      await page.goto('/');
      await goToCompressTab(page);
      const row = page.locator('tr:has-text("items.bin.huffman.compressed")');
      await expect(row.locator('.algorithm-badge:has-text("HUFFMAN")')).toBeVisible();
    });

    test('decompress items.bin.huffman.compressed', async ({ page }) => {
      await page.goto('/');
      await decompressFile(page, 'items.bin.huffman.compressed');
    });

    test('verify data integrity after Huffman decompress - first item', async ({ page }) => {
      await page.goto('/');
      const item = getSeededItem(0);
      await verifyItemExists(page, item.id, item.name);
    });

    test('verify data integrity after Huffman decompress - last item', async ({ page }) => {
      await page.goto('/');
      const item = getSeededItem(39); // Assuming 40 items in seed
      await verifyItemExists(page, item.id, item.name);
    });

    test('verify compressed file removed after decompress', async ({ page }) => {
      // Decompression removes the compressed file automatically
      await page.goto('/');
      await verifyCompressedFileNotExists(page, 'items.bin.huffman.compressed');
    });
  });

  test.describe('LZW Compression', () => {
    test('compress orders.bin with LZW', async ({ page }) => {
      await page.goto('/');
      await compressFile(page, 'orders.bin', 'lzw');
    });

    test('verify orders.bin.lzw.compressed exists', async ({ page }) => {
      await page.goto('/');
      await verifyCompressedFileExists(page, 'orders.bin.lzw.compressed');
    });

    test('compressed file shows correct algorithm', async ({ page }) => {
      await page.goto('/');
      await goToCompressTab(page);
      const row = page.locator('tr:has-text("orders.bin.lzw.compressed")');
      await expect(row.locator('.algorithm-badge:has-text("LZW")')).toBeVisible();
    });

    test('decompress orders.bin.lzw.compressed', async ({ page }) => {
      await page.goto('/');
      await decompressFile(page, 'orders.bin.lzw.compressed');
    });

    test('verify data integrity after LZW decompress - first order', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(0);
      await verifyOrderExists(page, order.id, order.customer);
    });

    test('verify data integrity after LZW decompress - last order', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(9); // Assuming 10 orders in seed
      await verifyOrderExists(page, order.id, order.customer);
    });

    test('verify compressed file removed after decompress', async ({ page }) => {
      // Decompression removes the compressed file automatically
      await page.goto('/');
      await verifyCompressedFileNotExists(page, 'orders.bin.lzw.compressed');
    });
  });

  test.describe('Compress All Files', () => {
    test('compress all files with Huffman', async ({ page }) => {
      await page.goto('/');
      await compressFile(page, '__all__', 'huffman');
    });

    test('verify combined compressed file exists', async ({ page }) => {
      await page.goto('/');
      // "All Files" compression creates a single combined file
      await verifyCompressedFileExists(page, 'all_files.huffman.compressed');
    });

    test('compressed file shows correct algorithm', async ({ page }) => {
      await page.goto('/');
      await goToCompressTab(page);
      const row = page.locator('tr:has-text("all_files.huffman.compressed")');
      await expect(row.locator('.algorithm-badge:has-text("HUFFMAN")')).toBeVisible();
    });

    test('decompress all files and verify data integrity', async ({ page }) => {
      await page.goto('/');
      // Decompress the combined file
      await decompressFile(page, 'all_files.huffman.compressed');
      // Verify data is intact - check one record from each entity type
      const item = getSeededItem(0);
      await verifyItemExists(page, item.id, item.name);
    });

    test('verify all entity data restored after decompress', async ({ page }) => {
      await page.goto('/');
      const order = getSeededOrder(0);
      await verifyOrderExists(page, order.id, order.customer);
      const promo = getSeededPromotion(0);
      await verifyPromotionExists(page, promo.id, promo.name);
    });
  });

  test.describe('Cleanup', () => {
    test('delete all files', async ({ page }) => {
      await page.goto('/');
      await deleteAllFiles(page);
    });

    test('verify no compressed files remain', async ({ page }) => {
      await page.goto('/');
      const count = await getCompressedFilesCount(page);
      expect(count).toBe(0);
    });
  });
});
