// @ts-check
import { test, expect } from '@playwright/test';

test.describe('Scripts Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/scripts');
    await expect(page.locator('h2')).toContainText('Script Management');
  });

  test('should display add new script button', async ({ page }) => {
    const addButton = page.locator('button:has-text("Add New Script")');
    await expect(addButton).toBeVisible();
    await expect(addButton).toBeEnabled();
  });

  test('should display existing scripts with complete information', async ({ page }) => {
    // Wait for scripts to load
    await page.waitForResponse(response => 
      response.url().includes('/api/scripts') && response.status() === 200
    );

    // Should have script cards
    const scriptCards = page.locator('[data-testid="script-card"]');
    const scriptCount = await scriptCards.count();
    
    if (scriptCount > 0) {
      // Each script card should have complete information
      for (let i = 0; i < scriptCount; i++) {
        const card = scriptCards.nth(i);
        
        // Should have script name
        const scriptName = card.locator('[data-testid="script-name"]');
        await expect(scriptName).toBeVisible();
        await expect(scriptName).not.toBeEmpty();
        
        // Should have script path
        const scriptPath = card.locator('[data-testid="script-path"]');
        await expect(scriptPath).toBeVisible();
        await expect(scriptPath).not.toBeEmpty();
        
        // Should have interval information
        const intervalInfo = card.locator('[data-testid="script-interval"]');
        await expect(intervalInfo).toBeVisible();
        await expect(intervalInfo).not.toContainText('Interval: s'); // Should not be empty
        
        // Should have status information
        const statusInfo = card.locator('[data-testid="script-status"]');
        await expect(statusInfo).toBeVisible();
        await expect(statusInfo).toContainText(/enabled|disabled/i);
        
        // Should have action buttons
        await expect(card.locator('button:has-text("Run Now")')).toBeVisible();
        await expect(card.locator('button:has-text("Edit")')).toBeVisible();
        await expect(card.locator('button:has-text("Delete")')).toBeVisible();
      }
    }
  });

  test('should handle empty scripts state', async ({ page }) => {
    // If no scripts, should show appropriate message
    const scriptCards = page.locator('[data-testid="script-card"]');
    const scriptCount = await scriptCards.count();
    
    if (scriptCount === 0) {
      const emptyMessage = page.locator('[data-testid="no-scripts-message"]');
      await expect(emptyMessage).toBeVisible();
      await expect(emptyMessage).toContainText(/no scripts/i);
    }
  });

  test('should load scripts from API successfully', async ({ page }) => {
    let apiCallSuccessful = false;
    let apiResponseData = null;

    page.on('response', async response => {
      if (response.url().includes('/api/scripts')) {
        apiCallSuccessful = true;
        expect(response.status()).toBe(200);
        try {
          apiResponseData = await response.json();
          expect(apiResponseData).toHaveProperty('success', true);
          expect(apiResponseData).toHaveProperty('data');
          expect(Array.isArray(apiResponseData.data)).toBe(true);
        } catch (error) {
          console.error('Failed to parse API response:', error);
        }
      }
    });

    await page.reload();
    await page.waitForTimeout(2000);
    
    expect(apiCallSuccessful).toBe(true);
  });

  test('should enable/disable scripts', async ({ page }) => {
    await page.waitForResponse(response => 
      response.url().includes('/api/scripts') && response.status() === 200
    );

    const scriptCards = page.locator('[data-testid="script-card"]');
    const scriptCount = await scriptCards.count();
    
    if (scriptCount > 0) {
      const firstCard = scriptCards.first();
      const enableButton = firstCard.locator('button:has-text("Enable")');
      const disableButton = firstCard.locator('button:has-text("Disable")');
      
      // Should have either enable or disable button (not both)
      const enableVisible = await enableButton.isVisible();
      const disableVisible = await disableButton.isVisible();
      
      expect(enableVisible || disableVisible).toBe(true);
      expect(enableVisible && disableVisible).toBe(false);
    }
  });
});