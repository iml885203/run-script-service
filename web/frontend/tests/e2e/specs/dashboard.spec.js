// @ts-check
import { test, expect } from '@playwright/test';

test.describe('Dashboard Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    // Wait for the page to load
    await expect(page.locator('h2')).toContainText('System Dashboard');
  });

  test('should display system status data', async ({ page }) => {
    // Test: System Status card should show actual status data
    const systemStatusCard = page.locator('[data-testid="system-status-card"]');
    await expect(systemStatusCard).toBeVisible();

    // Should not be empty - should contain actual status text
    const statusText = systemStatusCard.locator('[data-testid="status-value"]');
    await expect(statusText).not.toBeEmpty();
    await expect(statusText).toContainText(/running|stopped|error/i);
  });

  test('should display uptime data', async ({ page }) => {
    // Test: Uptime card should show actual uptime
    const uptimeCard = page.locator('[data-testid="uptime-card"]');
    await expect(uptimeCard).toBeVisible();

    const uptimeValue = uptimeCard.locator('[data-testid="uptime-value"]');
    await expect(uptimeValue).not.toBeEmpty();
    // Should contain time format (e.g., "2h 30m", "1d 5h", etc.)
    const uptimeText = await uptimeValue.textContent();
    expect(uptimeText).toMatch(/\d+[dhms]/);
  });

  test('should display running scripts count', async ({ page }) => {
    // Test: Running Scripts card should show count
    const runningScriptsCard = page.locator('[data-testid="running-scripts-card"]');
    await expect(runningScriptsCard).toBeVisible();

    const runningCount = runningScriptsCard.locator('[data-testid="running-scripts-value"]');
    await expect(runningCount).not.toBeEmpty();
    // Should be a number (0 or more)
    const runningCountText = await runningCount.textContent();
    expect(runningCountText).toMatch(/^\d+$/);
  });

  test('should display total scripts count', async ({ page }) => {
    // Test: Total Scripts card should show count
    const totalScriptsCard = page.locator('[data-testid="total-scripts-card"]');
    await expect(totalScriptsCard).toBeVisible();

    const totalCount = totalScriptsCard.locator('[data-testid="total-scripts-value"]');
    await expect(totalCount).not.toBeEmpty();
    // Should be a number (0 or more)
    const totalCountText = await totalCount.textContent();
    expect(totalCountText).toMatch(/^\d+$/);
  });

  test('should load data from API', async ({ page }) => {
    // Test: Page should make API calls and receive data
    let statusApiCalled = false;
    let scriptsApiCalled = false;

    page.on('response', response => {
      if (response.url().includes('/api/status')) {
        statusApiCalled = true;
        expect(response.status()).toBe(200);
      }
      if (response.url().includes('/api/scripts')) {
        scriptsApiCalled = true;
        expect(response.status()).toBe(200);
      }
    });

    await page.reload();

    // Wait for API calls to complete
    await page.waitForTimeout(2000);

    expect(statusApiCalled).toBe(true);
    expect(scriptsApiCalled).toBe(true);
  });

  test('should have working navigation to scripts', async ({ page }) => {
    // Test: Scripts overview link should work - use the specific "Add scripts" link
    const addScriptsLink = page.getByRole('link', { name: 'Add scripts' });
    await expect(addScriptsLink).toBeVisible();

    await addScriptsLink.click();
    await expect(page).toHaveURL('/scripts');
    await expect(page.locator('h2')).toContainText('Script Management');
  });
});
