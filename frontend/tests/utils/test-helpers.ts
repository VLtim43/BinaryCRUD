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

// Seed data loader - tests run from frontend/, seed is at ../data/seed/
const SEED_DIR = path.resolve(process.cwd(), '../data/seed');

function loadSeedFile<T>(filename: string): T[] {
  const filePath = path.join(SEED_DIR, filename);
  const content = fs.readFileSync(filePath, 'utf-8');
  return JSON.parse(content);
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

// Random selection helpers
export function getRandomSeededItem(): { id: number; name: string; priceInCents: number } {
  const items = getSeedItems();
  const index = Math.floor(Math.random() * items.length);
  return { id: index, ...items[index] };
}

export function getRandomSeededOrder(): { id: number; customer: string } {
  const orders = getSeedOrders();
  const index = Math.floor(Math.random() * orders.length);
  return { id: index, customer: orders[index].owner };
}

export function getRandomSeededPromotion(): { id: number; name: string } {
  const promotions = getSeedPromotions();
  const index = Math.floor(Math.random() * promotions.length);
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
  await input.fill('');
  await input.type(value);
  // Wait for React state to update
  await page.waitForTimeout(100);
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
  await expect(page.locator(`.toast:has-text("Item saved")`)).toBeVisible({ timeout: 5000 });
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
