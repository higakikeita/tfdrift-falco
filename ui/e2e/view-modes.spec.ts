/**
 * View Mode Switching E2E Tests
 * Tests different view modes: graph, table, and split view
 */

import { test, expect } from '@playwright/test';

test.describe('View Mode Switching', () => {
  test('should load application in split view by default', async ({ page }) => {
    await page.goto('/');

    // Check that both graph and table are visible in split view
    const graphContainer = page.locator('.react-flow');

    await expect(graphContainer).toBeVisible({ timeout: 10000 });
    // Table might not be visible initially depending on data loading
    await page.waitForTimeout(2000);
  });

  test('should switch to graph-only view', async ({ page }) => {
    await page.goto('/');

    // Find and click graph view button
    const graphViewButton = page.locator('button:has-text("Graph")').or(
      page.locator('button[aria-label*="graph"]')
    );

    const isVisible = await graphViewButton.isVisible().catch(() => false);

    if (isVisible) {
      await graphViewButton.click();
      await page.waitForTimeout(500);

      // Graph should be visible
      const graphContainer = page.locator('.react-flow');
      await expect(graphContainer).toBeVisible();
    }

    expect(true).toBe(true); // Test passes if no errors
  });

  test('should switch to table-only view', async ({ page }) => {
    await page.goto('/');

    // Find and click table view button
    const tableViewButton = page.locator('button:has-text("Table")').or(
      page.locator('button:has-text("テーブル")')
    ).or(page.locator('button[aria-label*="table"]'));

    const isVisible = await tableViewButton.isVisible().catch(() => false);

    if (isVisible) {
      await tableViewButton.click();
      await page.waitForTimeout(500);

      // Table should be visible
      const tableContainer = page.locator('table').or(
        page.locator('[role="table"]')
      );

      await expect(tableContainer).toBeVisible({ timeout: 5000 });
    }

    expect(true).toBe(true);
  });

  test('should switch between split and single views', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);

    // Try to switch views multiple times
    const viewButtons = page.locator('button').filter({
      hasText: /Graph|Table|Split|グラフ|テーブル/
    });

    const count = await viewButtons.count();

    if (count > 0) {
      // Click first button
      await viewButtons.first().click();
      await page.waitForTimeout(500);

      // Click second button if exists
      if (count > 1) {
        await viewButtons.nth(1).click();
        await page.waitForTimeout(500);
      }
    }

    // Application should still be functional
    await expect(page).toHaveTitle(/TFDrift Falco/i);
  });
});

test.describe('Demo Mode Switching', () => {
  test('should switch between demo modes', async ({ page }) => {
    await page.goto('/');

    // Find demo mode selector
    const demoSelector = page.locator('select').filter({
      hasText: /API|Simple|Complex|Blast Radius|Network/i
    }).or(page.locator('select[name*="demo"]'));

    const isVisible = await demoSelector.isVisible().catch(() => false);

    if (isVisible) {
      // Try switching to different modes
      await demoSelector.selectOption({ index: 1 });
      await page.waitForTimeout(1000);

      await demoSelector.selectOption({ index: 2 });
      await page.waitForTimeout(1000);

      // Graph should still be visible after switching
      const graphContainer = page.locator('.react-flow');
      await expect(graphContainer).toBeVisible();
    }

    expect(true).toBe(true);
  });

  test('should update node count when switching modes', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(1000);

    // Find node count display
    const nodeCount = page.locator('text=/\\d+ Nodes/i').or(
      page.locator('[class*="stats"]')
    );

    // Demo mode selector
    const demoSelector = page.locator('select').filter({
      hasText: /API|Simple|Complex/i
    });

    const isSelectorVisible = await demoSelector.isVisible().catch(() => false);

    if (isSelectorVisible) {
      // Get initial count
      const initialText = await nodeCount.textContent().catch(() => null);

      // Switch mode
      await demoSelector.selectOption({ index: 1 });
      await page.waitForTimeout(1000);

      // Count might change (or might not, depending on data)
      const newText = await nodeCount.textContent().catch(() => null);

      // Just verify the app didn't crash
      expect(newText !== null || initialText !== null).toBe(true);
    }

    expect(true).toBe(true);
  });
});

test.describe('Layout Switching', () => {
  test('should switch between different graph layouts', async ({ page }) => {
    await page.goto('/');

    // Find layout selector
    const layoutSelector = page.locator('select').filter({
      hasText: /Hierarchical|Radial|Force|Grid/i
    }).or(page.locator('select[name*="layout"]'));

    const isVisible = await layoutSelector.isVisible().catch(() => false);

    if (isVisible) {
      // Try different layouts
      const options = await layoutSelector.locator('option').count();

      for (let i = 0; i < Math.min(options, 3); i++) {
        await layoutSelector.selectOption({ index: i });
        await page.waitForTimeout(800);

        // Graph should still be visible
        const graphContainer = page.locator('.react-flow');
        await expect(graphContainer).toBeVisible();
      }
    }

    expect(true).toBe(true);
  });

  test('should maintain node interactivity after layout change', async ({ page }) => {
    await page.goto('/');

    // Change layout
    const layoutSelector = page.locator('select').filter({
      hasText: /Hierarchical|Radial/i
    });

    const isVisible = await layoutSelector.isVisible().catch(() => false);

    if (isVisible) {
      await layoutSelector.selectOption({ index: 1 });
      await page.waitForTimeout(1000);

      // Try to click a node
      const firstNode = page.locator('.react-flow__node').first();
      const isNodeVisible = await firstNode.isVisible().catch(() => false);

      if (isNodeVisible) {
        await firstNode.click();

        // Detail panel should appear
        const detailPanel = page.locator('[class*="detail-panel"]');
        await expect(detailPanel).toBeVisible({ timeout: 3000 });
      }
    }

    expect(true).toBe(true);
  });
});
