// @ts-check
import { test, expect } from '@playwright/test';

test.describe('Scripts Page - Simple Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to scripts page
    await page.goto('/scripts');
  });

  test('should load scripts page with title', async ({ page }) => {
    // Check page title
    await expect(page.locator('h2')).toHaveText('Script Management');
  });

  test('should have add new script button', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add New Script")');
    await expect(addButton).toBeVisible();
    await expect(addButton).toBeEnabled();
  });

  test('should check if API call is being made', async ({ page }) => {
    let apiCallMade = false;
    
    // Listen for API requests
    page.on('request', request => {
      if (request.url().includes('/api/scripts')) {
        console.log('API request detected:', request.url());
        apiCallMade = true;
      }
    });

    page.on('response', response => {
      if (response.url().includes('/api/scripts')) {
        console.log('API response received:', response.status(), response.url());
      }
    });

    // Wait a moment for any API calls to be made
    await page.waitForTimeout(3000);

    // Log whether API call was made
    console.log('API call made:', apiCallMade);
  });

  test('should display some content (either scripts or loading)', async ({ page }) => {
    // Check if page shows loading, error, or script content
    const hasLoading = await page.locator('.loading').isVisible();
    const hasError = await page.locator('.error').isVisible();
    const hasScripts = await page.locator('[data-testid="script-card"]').count() > 0;
    const hasNoScriptsMessage = await page.locator('[data-testid="no-scripts-message"]').isVisible();

    console.log('Page state:', { hasLoading, hasError, hasScripts, hasNoScriptsMessage });

    // At least one of these should be true
    expect(hasLoading || hasError || hasScripts || hasNoScriptsMessage).toBe(true);
  });

  test('should check JavaScript console for errors', async ({ page }) => {
    const consoleMessages = [];
    
    page.on('console', msg => {
      consoleMessages.push({
        type: msg.type(),
        text: msg.text()
      });
    });

    // Wait for page to load
    await page.waitForTimeout(2000);

    // Log all console messages
    console.log('Console messages:', consoleMessages);

    // Check for JavaScript errors
    const errors = consoleMessages.filter(msg => msg.type === 'error');
    if (errors.length > 0) {
      console.log('JavaScript errors found:', errors);
    }
  });
});