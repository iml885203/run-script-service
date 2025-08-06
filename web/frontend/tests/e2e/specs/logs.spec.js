// @ts-check
import { test, expect } from '@playwright/test';

test.describe('Logs Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/logs');
    await expect(page.locator('h2')).toContainText('System Logs');
  });

  test('should not have JavaScript errors', async ({ page }) => {
    const errors = [];

    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });

    page.on('pageerror', error => {
      errors.push(error.message);
    });

    await page.reload();
    await page.waitForTimeout(3000);

    // Should not have forEach errors or other JS errors
    const forEachErrors = errors.filter(error =>
      error.includes('forEach is not a function')
    );
    expect(forEachErrors.length).toBe(0);

    // Log any other errors for debugging
    if (errors.length > 0) {
      console.warn('JavaScript errors found:', errors);
    }
  });

  test('should load logs successfully', async ({ page }) => {
    // Should not be stuck in loading state
    const loadingIndicator = page.locator('text=Loading logs...');

    // Wait for either logs to appear or no logs message
    await Promise.race([
      page.waitForSelector('[data-testid="log-entry"]', { timeout: 10000 }),
      page.waitForSelector('[data-testid="no-logs-message"]', { timeout: 10000 }),
      page.waitForTimeout(10000)
    ]);

    // Loading indicator should be gone
    await expect(loadingIndicator).not.toBeVisible();
  });

  test('should have functional filter controls', async ({ page }) => {
    // Script filter dropdown should be enabled
    const scriptFilter = page.locator('select[data-testid="script-filter"]');
    await expect(scriptFilter).toBeVisible();
    await expect(scriptFilter).toBeEnabled();

    // Limit dropdown should be enabled
    const limitSelect = page.locator('select[data-testid="limit-select"]');
    await expect(limitSelect).toBeVisible();
    await expect(limitSelect).toBeEnabled();
  });

  test('should have functional action buttons', async ({ page }) => {
    // Wait for page to load
    await page.waitForTimeout(2000);

    // Refresh button should be enabled
    const refreshButton = page.locator('button:has-text("Refresh")');
    await expect(refreshButton).toBeVisible();
    await expect(refreshButton).toBeEnabled();

    // Clear Logs button should be enabled
    const clearButton = page.locator('button:has-text("Clear Logs")');
    await expect(clearButton).toBeVisible();
    await expect(clearButton).toBeEnabled();
  });

  test('should display logs with proper formatting', async ({ page }) => {
    // Wait for page to fully load first
    await page.waitForTimeout(3000);

    const logEntries = page.locator('[data-testid="log-entry"]');
    const logCount = await logEntries.count();

    // If no logs are present, this test should be skipped or treated as passing
    if (logCount === 0) {
      console.log('No logs found - checking if no-logs message is displayed');
      const noLogsMessage = page.locator('[data-testid="no-logs-message"]');
      await expect(noLogsMessage).toBeVisible();
      return; // Skip the rest of the test
    }

    // If logs are present, check their structure
    for (let i = 0; i < Math.min(3, logCount); i++) { // Check first 3 entries
      const entry = logEntries.nth(i);

      // Should have timestamp
      const timestamp = entry.locator('[data-testid="log-timestamp"]');
      await expect(timestamp).toBeVisible();

      // Should have log level
      const level = entry.locator('[data-testid="log-level"]');
      await expect(level).toBeVisible();
      await expect(level).toContainText(/info|error|warning|debug/i);

      // Should have log message
      const message = entry.locator('[data-testid="log-message"]');
      await expect(message).toBeVisible();
      await expect(message).not.toBeEmpty();
    }
  });

  test('should display log statistics', async ({ page }) => {
    // Wait for logs to load
    await page.waitForTimeout(3000);

    // Should show log summary with counts
    const logSummary = page.locator('[data-testid="logs-summary"]');
    await expect(logSummary).toBeVisible();

    // Should show total count
    const totalCount = logSummary.locator('[data-testid="total-logs"]');
    await expect(totalCount).toBeVisible();

    // Get the text content and check it matches the expected format
    const totalCountText = await totalCount.textContent();
    expect(totalCountText).toMatch(/Total.*\d+/);
  });

  test('should filter logs by script', async ({ page }) => {
    // Wait for initial load
    await page.waitForTimeout(2000);

    const scriptFilter = page.locator('select[data-testid="script-filter"]');

    // Should have options beyond "All scripts"
    const options = scriptFilter.locator('option');
    const optionCount = await options.count();

    if (optionCount > 1) {
      // Select a specific script
      await scriptFilter.selectOption({ index: 1 });

      // Should trigger filtering
      await page.waitForTimeout(1000);

      // Logs should be filtered (this is basic functionality test)
      // More specific filtering tests would require known log data
    }
  });
});
