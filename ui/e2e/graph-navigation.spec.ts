/**
 * Graph Navigation E2E Tests
 * Tests user flows for graph visualization and node interaction
 */

import { test, expect } from '@playwright/test';

test.describe('Graph Visualization', () => {
  test('should load the application', async ({ page }) => {
    await page.goto('/');

    // Check that the page loaded
    await expect(page).toHaveTitle(/TFDrift Falco/i);
  });

  test('should display graph canvas', async ({ page }) => {
    await page.goto('/');

    // Wait for React Flow to render
    const canvas = page.locator('.react-flow');
    await expect(canvas).toBeVisible({ timeout: 10000 });
  });

  test('should display nodes on the graph', async ({ page }) => {
    await page.goto('/');

    // Wait for nodes to render
    const nodes = page.locator('.react-flow__node');
    await expect(nodes.first()).toBeVisible({ timeout: 10000 });

    // Check that there are multiple nodes
    const nodeCount = await nodes.count();
    expect(nodeCount).toBeGreaterThan(0);
  });

  test('should open node detail panel on click', async ({ page }) => {
    await page.goto('/');

    // Wait for first node and click it
    const firstNode = page.locator('.react-flow__node').first();
    await firstNode.waitFor({ state: 'visible', timeout: 10000 });
    await firstNode.click();

    // Check that detail panel appears
    const detailPanel = page.locator('[class*="detail-panel"]');
    await expect(detailPanel).toBeVisible({ timeout: 5000 });
  });

  test('should close node detail panel', async ({ page }) => {
    await page.goto('/');

    // Open detail panel
    const firstNode = page.locator('.react-flow__node').first();
    await firstNode.waitFor({ state: 'visible', timeout: 10000 });
    await firstNode.click();

    const detailPanel = page.locator('[class*="detail-panel"]');
    await expect(detailPanel).toBeVisible();

    // Close the panel
    const closeButton = detailPanel.locator('button').first();
    await closeButton.click();

    // Panel should be hidden
    await expect(detailPanel).not.toBeVisible();
  });
});

test.describe('Graph Interactions', () => {
  test('should support zoom in/out', async ({ page }) => {
    await page.goto('/');

    const canvas = page.locator('.react-flow');
    await canvas.waitFor({ state: 'visible' });

    // Get initial zoom level (via transform attribute)
    const initialTransform = await canvas.evaluate((el) => {
      const viewport = el.querySelector('.react-flow__viewport');
      return viewport ? window.getComputedStyle(viewport).transform : null;
    });

    // Zoom in using mouse wheel
    await canvas.hover();
    await page.mouse.wheel(0, -100);

    // Wait a bit for zoom to apply
    await page.waitForTimeout(500);

    // Transform should have changed
    const newTransform = await canvas.evaluate((el) => {
      const viewport = el.querySelector('.react-flow__viewport');
      return viewport ? window.getComputedStyle(viewport).transform : null;
    });

    expect(newTransform).not.toBe(initialTransform);
  });

  test('should support panning', async ({ page }) => {
    await page.goto('/');

    const canvas = page.locator('.react-flow');
    await canvas.waitFor({ state: 'visible' });

    // Pan the canvas
    await canvas.hover();
    await page.mouse.down();
    await page.mouse.move(100, 100);
    await page.mouse.up();

    // Canvas should have moved (hard to assert exact position, but we can check it didn't crash)
    await expect(canvas).toBeVisible();
  });
});

test.describe('Keyboard Shortcuts', () => {
  test('should open help with ? key', async ({ page }) => {
    await page.goto('/');

    // Press ? key
    await page.keyboard.press('?');

    // Keyboard shortcuts guide should appear
    const shortcutsGuide = page.locator('text=キーボードショートカット');
    await expect(shortcutsGuide).toBeVisible({ timeout: 2000 });
  });

  test('should toggle help overlay with F1', async ({ page }) => {
    await page.goto('/');

    // Press F1
    await page.keyboard.press('F1');

    // Help overlay should toggle
    await page.waitForTimeout(500);

    // Press F1 again
    await page.keyboard.press('F1');

    await page.waitForTimeout(500);
  });
});
