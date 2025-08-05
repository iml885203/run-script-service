// @ts-check
import { test, expect } from '@playwright/test';

test.describe('Scripts Page - Debug', () => {
  test('should debug the exact state of the page', async ({ page }) => {
    await page.goto('/scripts');
    
    // Wait for the page to load completely
    await page.waitForLoadState('networkidle');
    
    // Take a screenshot for debugging
    await page.screenshot({ path: 'debug-scripts-page.png' });
    
    // Log the entire page content
    const pageContent = await page.content();
    console.log('Full page HTML:', pageContent.substring(0, 2000) + '...');
    
    // Check for scripts API call
    let apiResponse = null;
    page.on('response', async response => {
      if (response.url().includes('/api/scripts')) {
        apiResponse = await response.json();
        console.log('API Response:', JSON.stringify(apiResponse, null, 2));
      }
    });
    
    // Wait longer for API call
    await page.waitForTimeout(5000);
    
    // Check what elements exist
    const scriptCards = await page.locator('[data-testid="script-card"]').count();
    const hasLoading = await page.locator('.loading').isVisible();
    const hasError = await page.locator('.error').isVisible();
    const hasNoScripts = await page.locator('[data-testid="no-scripts-message"]').isVisible();
    
    console.log('Page state:', {
      scriptCards,
      hasLoading,
      hasError,
      hasNoScripts,
      apiResponse
    });
    
    // Get text content of script cards if any exist
    if (scriptCards > 0) {
      for (let i = 0; i < scriptCards; i++) {
        const card = page.locator('[data-testid="script-card"]').nth(i);
        const name = await card.locator('[data-testid="script-name"]').textContent();
        const path = await card.locator('[data-testid="script-path"]').textContent();
        const interval = await card.locator('[data-testid="script-interval"]').textContent();
        const status = await card.locator('[data-testid="script-status"]').textContent();
        
        console.log(`Script ${i}:`, { name, path, interval, status });
      }
    }
    
    // Get the body text to see what's actually displayed
    const bodyText = await page.locator('body').textContent();
    console.log('Body text:', bodyText);
  });
});