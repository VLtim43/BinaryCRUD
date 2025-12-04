/// <reference types="node" />
import { Page, expect } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

// Seed data types
interface SeedItem {
  name: string;
  priceInCents: number;
}

interface SeedOrder {
  owner: string;
  itemIDs: number[];
  promotionIDs?: number[];
}

interface SeedPromotion {
  name: string;
  itemIDs: number[];
}

// Seed data loader with caching - tests run from frontend/, seed is at ../data/seed/
const SEED_DIR = path.resolve(process.cwd(), '../data/seed');

// Cache seed data to avoid repeated disk reads
const seedCache: Record<string, unknown[]> = {};

function loadSeedFile<T>(filename: string): T[] {
  if (!seedCache[filename]) {
    const filePath = path.join(SEED_DIR, filename);
    const content = fs.readFileSync(filePath, 'utf-8');
    seedCache[filename] = JSON.parse(content);
  }
  return seedCache[filename] as T[];
}

export function getSeedItems(): SeedItem[] {
  return loadSeedFile<SeedItem>('items.json');
}

export function getSeedOrders(): SeedOrder[] {
  return loadSeedFile<SeedOrder>('orders.json');
}

export function getSeedPromotions(): SeedPromotion[] {
  return loadSeedFile<SeedPromotion>('promotions.json');
}

// Get item with its expected ID (0-indexed after populate)
export function getSeededItem(index: number): { id: number; name: string; priceInCents: number } {
  const items = getSeedItems();
  return { id: index, ...items[index] };
}

export function getSeededOrder(index: number): { id: number; customer: string } {
  const orders = getSeedOrders();
  return { id: index, customer: orders[index].owner };
}

export function getSeededPromotion(index: number): { id: number; name: string } {
  const promotions = getSeedPromotions();
  return { id: index, name: promotions[index].name };
}

// Tab navigation helpers
export async function goToTab(page: Page, tabName: 'Item' | 'Order' | 'Promotion' | 'Debug') {
  await page.click(`button.tab:has-text("${tabName}")`);
}

export async function goToSubTab(page: Page, subTabName: 'Create' | 'Read' | 'Delete') {
  await page.click(`.sub_tabs button:has-text("${subTabName}")`);
}

// Helper to fill controlled inputs reliably
async function fillInput(page: Page, selector: string, value: string) {
  const input = page.locator(selector);
  await input.click();
  await input.fill(value);
}

// Item CRUD helpers
export async function readItem(page: Page, id: number) {
  await goToTab(page, 'Item');
  await goToSubTab(page, 'Read');
  await fillInput(page, '#record-id', id.toString());
  await page.click('button:has-text("Get Record")');
}

export async function verifyItemExists(page: Page, id: number, expectedName?: string) {
  await readItem(page, id);
  await expect(page.locator('.details-card')).toBeVisible();
  await expect(page.locator('.details-value').first()).toHaveText(id.toString());
  if (expectedName) {
    await expect(page.locator('.details-row:has(.details-label:has-text("Name")) .details-value')).toHaveText(expectedName);
  }
}

export async function verifyItemNotFound(page: Page, id: number) {
  await readItem(page, id);
  await expect(page.locator('.details-card')).not.toBeVisible();
  // Should show error toast
  await expect(page.locator('.toast-error, .toast:has-text("not found"), .toast:has-text("error")')).toBeVisible({ timeout: 5000 });
}

export async function deleteItem(page: Page, id: number) {
  await goToTab(page, 'Item');
  await goToSubTab(page, 'Delete');
  await fillInput(page, '#delete-record-id', id.toString());
  await page.click('button:has-text("Delete Record")');
  await expect(page.locator(`.toast:has-text("Item ${id} deleted")`)).toBeVisible({ timeout: 5000 });
}

export async function createItem(page: Page, name: string, price: string) {
  await goToTab(page, 'Item');
  await goToSubTab(page, 'Create');
  await fillInput(page, '#name', name);
  await fillInput(page, '#price', price);
  await page.click('button:has-text("Add Item")');
  // Toast: "Item saved: {name} (${price})"
  await expect(page.locator(`.toast:has-text("Item saved")`)).toBeVisible({ timeout: 5000 });
}

// Order creation is more complex - requires adding items to cart first
// This is a simplified version that assumes items are already in cart
export async function createOrder(page: Page, customerName: string) {
  await goToTab(page, 'Order');
  await goToSubTab(page, 'Create');
  await fillInput(page, '#customer-name', customerName);
  await page.click('button:has-text("Create Order")');
  // Toast: "Order #{id} created for {customerName}"
  await expect(page.locator(`.toast:has-text("Order #")`)).toBeVisible({ timeout: 5000 });
}

export async function createPromotion(page: Page, promotionName: string) {
  await goToTab(page, 'Promotion');
  await goToSubTab(page, 'Create');
  await fillInput(page, '#promotion-name', promotionName);
  await page.click('button:has-text("Create Promotion")');
  // Toast: "Promotion #{id} created: {promotionName}"
  await expect(page.locator(`.toast:has-text("Promotion #")`)).toBeVisible({ timeout: 5000 });
}

// Order CRUD helpers
export async function readOrder(page: Page, id: number) {
  await goToTab(page, 'Order');
  await goToSubTab(page, 'Read');
  await fillInput(page, '#record-id', id.toString());
  await page.click('button:has-text("Get Record")');
}

export async function verifyOrderExists(page: Page, id: number, expectedCustomer?: string) {
  await readOrder(page, id);
  await expect(page.locator('.details-card')).toBeVisible();
  await expect(page.locator('.details-value').first()).toHaveText(id.toString());
  if (expectedCustomer) {
    await expect(page.locator('.details-row:has(.details-label:has-text("Customer")) .details-value')).toHaveText(expectedCustomer);
  }
}

export async function verifyOrderNotFound(page: Page, id: number) {
  await readOrder(page, id);
  await expect(page.locator('.details-card')).not.toBeVisible();
  await expect(page.locator('.toast-error, .toast:has-text("not found"), .toast:has-text("error")')).toBeVisible({ timeout: 5000 });
}

export async function deleteOrder(page: Page, id: number) {
  await goToTab(page, 'Order');
  await goToSubTab(page, 'Delete');
  await fillInput(page, '#delete-record-id', id.toString());
  await page.click('button:has-text("Delete Record")');
  await expect(page.locator(`.toast:has-text("Order ${id} deleted")`)).toBeVisible({ timeout: 5000 });
}

// Promotion CRUD helpers
export async function readPromotion(page: Page, id: number) {
  await goToTab(page, 'Promotion');
  await goToSubTab(page, 'Read');
  await fillInput(page, '#record-id', id.toString());
  await page.click('button:has-text("Get Record")');
}

export async function verifyPromotionExists(page: Page, id: number, expectedName?: string) {
  await readPromotion(page, id);
  await expect(page.locator('.details-card')).toBeVisible();
  await expect(page.locator('.details-value').first()).toHaveText(id.toString());
  if (expectedName) {
    await expect(page.locator('.details-row:has(.details-label:has-text("Name")) .details-value')).toHaveText(expectedName);
  }
}

export async function verifyPromotionNotFound(page: Page, id: number) {
  await readPromotion(page, id);
  await expect(page.locator('.details-card')).not.toBeVisible();
  await expect(page.locator('.toast-error, .toast:has-text("not found"), .toast:has-text("error")')).toBeVisible({ timeout: 5000 });
}

export async function deletePromotion(page: Page, id: number) {
  await goToTab(page, 'Promotion');
  await goToSubTab(page, 'Delete');
  await fillInput(page, '#delete-record-id', id.toString());
  await page.click('button:has-text("Delete Record")');
  await expect(page.locator(`.toast:has-text("Promotion ${id} deleted")`)).toBeVisible({ timeout: 5000 });
}

// Debug helpers
export async function deleteAllFiles(page: Page) {
  await goToTab(page, 'Debug');
  await page.click('button:has-text("Delete All Files")');
  await page.waitForTimeout(2000);
}

export async function populateInventory(page: Page) {
  await goToTab(page, 'Debug');
  await page.click('button:has-text("Populate Inventory")');
  await expect(page.locator('text=Inventory populated successfully')).toBeVisible({ timeout: 15000 });
}

// Generic test patterns
export async function testDeleteThenReadFails(
  page: Page,
  entityType: 'item' | 'order' | 'promotion',
  id: number
) {
  const deleteFunc = entityType === 'item' ? deleteItem : entityType === 'order' ? deleteOrder : deletePromotion;
  const verifyNotFoundFunc = entityType === 'item' ? verifyItemNotFound : entityType === 'order' ? verifyOrderNotFound : verifyPromotionNotFound;

  await deleteFunc(page, id);
  await verifyNotFoundFunc(page, id);
}

// Debug subtab navigation (Debug has different subtabs: Tools, Print, Compress)
export async function goToDebugSubTab(page: Page, subTabName: 'Tools' | 'Print' | 'Compress') {
  await goToTab(page, 'Debug');
  await page.click(`.sub_tabs button:has-text("${subTabName}")`);
}

// Compression helpers
export async function goToCompressTab(page: Page) {
  await goToDebugSubTab(page, 'Compress');
  await page.waitForTimeout(500);
}

export async function selectFileToCompress(page: Page, filename: string | '__all__') {
  await goToCompressTab(page);
  // First select dropdown is for file selection
  const fileSelect = page.locator('select').first();
  await fileSelect.selectOption(filename);
}

export async function selectCompressionAlgorithm(page: Page, algorithm: 'huffman' | 'lzw') {
  // Second select dropdown is for algorithm
  const algorithmSelect = page.locator('select').nth(1);
  await algorithmSelect.selectOption(algorithm);
}

export async function compressFile(page: Page, filename: string | '__all__', algorithm: 'huffman' | 'lzw') {
  await selectFileToCompress(page, filename);
  await selectCompressionAlgorithm(page, algorithm);
  // Wait for button to be enabled - use CSS :not() to exclude tab buttons
  const compressBtn = page.locator('button:has-text("Compress"):not(.tab)');
  await expect(compressBtn).toBeEnabled({ timeout: 5000 });
  await compressBtn.click();
  // Wait for compression to complete - toast says "Compressed: X saved" or "Compressed all files: X saved"
  await expect(page.locator('.toast:has-text("Compressed")')).toBeVisible({ timeout: 10000 });
}

export async function getCompressedFilesCount(page: Page): Promise<number> {
  await goToCompressTab(page);
  // Check if "Compressed Files" section exists
  const header = page.locator('h3:has-text("Compressed Files")');
  if (await header.isVisible()) {
    const text = await header.textContent();
    const match = text?.match(/\((\d+)\)/);
    return match ? parseInt(match[1], 10) : 0;
  }
  return 0;
}

export async function decompressFile(page: Page, filename: string) {
  await goToCompressTab(page);
  // Find the row with this filename and click Decompress
  const row = page.locator(`tr:has-text("${filename}")`);
  // First verify the row exists with a reasonable timeout
  await expect(row).toBeVisible({ timeout: 5000 });
  await row.locator('button:has-text("Decompress")').click();
  // Wait for decompression toast
  await expect(page.locator('.toast:has-text("Decompressed")')).toBeVisible({ timeout: 10000 });
}

export async function deleteCompressedFile(page: Page, filename: string) {
  await goToCompressTab(page);
  // Find the row with this filename and click Delete
  const row = page.locator(`tr:has-text("${filename}")`);
  // First verify the row exists with a reasonable timeout
  await expect(row).toBeVisible({ timeout: 5000 });
  await row.locator('button:has-text("Delete")').click();
  // Wait for deletion toast
  await expect(page.locator(`.toast:has-text("Deleted ${filename}")`)).toBeVisible({ timeout: 5000 });
}

export async function getBinFilesCount(page: Page): Promise<number> {
  await goToCompressTab(page);
  // Count options in the file select dropdown (minus "All Files" option)
  const fileSelect = page.locator('select').first();
  const optionCount = await fileSelect.locator('option').count();
  // Subtract 1 for "All Files" option
  return Math.max(0, optionCount - 1);
}

export async function verifyCompressedFileExists(page: Page, filename: string) {
  await goToCompressTab(page);
  await expect(page.locator(`tr:has-text("${filename}")`)).toBeVisible();
}

export async function verifyCompressedFileNotExists(page: Page, filename: string) {
  await goToCompressTab(page);
  await expect(page.locator(`tr:has-text("${filename}")`)).not.toBeVisible();
}
