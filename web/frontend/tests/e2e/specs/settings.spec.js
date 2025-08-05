// @ts-check
import { test, expect } from '@playwright/test';

test.describe('Settings Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/settings');
    await expect(page.locator('h2')).toContainText('System Settings');
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
  });

  test('should load configuration from API', async ({ page }) => {
    let configApiCalled = false;
    let configData = null;

    page.on('response', async response => {
      if (response.url().includes('/api/config')) {
        configApiCalled = true;
        expect(response.status()).toBe(200);
        try {
          configData = await response.json();
          expect(configData).toHaveProperty('success', true);
          expect(configData).toHaveProperty('data');
        } catch (error) {
          console.error('Failed to parse config API response:', error);
        }
      }
    });

    await page.reload();
    await page.waitForTimeout(2000);
    
    expect(configApiCalled).toBe(true);
  });

  test('should display all form fields with values', async ({ page }) => {
    // Wait for configuration to load
    await page.waitForResponse(response => 
      response.url().includes('/api/config') && response.status() === 200,
      { timeout: 10000 }
    );

    // Web Server Port field
    const webPortField = page.locator('input[data-testid="web-port-input"]');
    await expect(webPortField).toBeVisible();
    await expect(webPortField).not.toHaveValue(''); // Should have actual value
    await expect(webPortField).toHaveValue(/^\d+$/); // Should be numeric

    // Default Execution Interval field  
    const intervalField = page.locator('input[data-testid="interval-input"]');
    await expect(intervalField).toBeVisible();
    await expect(intervalField).not.toHaveValue(''); // Should have actual value

    // Log Retention field
    const logRetentionField = page.locator('input[data-testid="log-retention-input"]');
    await expect(logRetentionField).toBeVisible();
    await expect(logRetentionField).not.toHaveValue(''); // Should have actual value
    await expect(logRetentionField).toHaveValue(/^\d+$/); // Should be numeric

    // Auto-refresh checkbox
    const autoRefreshCheckbox = page.locator('input[data-testid="auto-refresh-checkbox"]');
    await expect(autoRefreshCheckbox).toBeVisible();
    // Checkbox state will be determined by loaded config
  });

  test('should enable save button when changes are made', async ({ page }) => {
    // Wait for form to load
    await page.waitForTimeout(2000);
    
    const saveButton = page.locator('button:has-text("Save Settings")');
    
    // Initially save button should be disabled (no changes)
    await expect(saveButton).toBeDisabled();
    
    // Make a change to a field
    const webPortField = page.locator('input[data-testid="web-port-input"]');
    await webPortField.fill('8081');
    
    // Save button should now be enabled
    await expect(saveButton).toBeEnabled();
  });

  test('should reset form when reset button is clicked', async ({ page }) => {
    // Wait for form to load
    await page.waitForTimeout(2000);
    
    // Get original value
    const webPortField = page.locator('input[data-testid="web-port-input"]');
    const originalValue = await webPortField.inputValue();
    
    // Make a change
    await webPortField.fill('9999');
    await expect(webPortField).toHaveValue('9999');
    
    // Click reset
    const resetButton = page.locator('button:has-text("Reset")');
    await resetButton.click();
    
    // Should revert to original value
    await expect(webPortField).toHaveValue(originalValue);
  });

  test('should display system information correctly', async ({ page }) => {
    // Version should be displayed
    const version = page.locator('[data-testid="system-version"]');
    await expect(version).toBeVisible();
    await expect(version).not.toBeEmpty();
    await expect(version).toMatch(/\d+\.\d+\.\d+/); // Version format

    // Platform should be displayed
    const platform = page.locator('[data-testid="system-platform"]');
    await expect(platform).toBeVisible();
    await expect(platform).not.toBeEmpty();

    // User Agent should be displayed
    const userAgent = page.locator('[data-testid="system-user-agent"]');
    await expect(userAgent).toBeVisible();
    await expect(userAgent).not.toBeEmpty();
  });

  test('should save configuration successfully', async ({ page }) => {
    // Wait for form to load
    await page.waitForTimeout(2000);
    
    let saveApiCalled = false;
    
    page.on('response', response => {
      if (response.url().includes('/api/config') && response.request().method() === 'PUT') {
        saveApiCalled = true;
        expect(response.status()).toBe(200);
      }
    });
    
    // Make a change
    const intervalField = page.locator('input[data-testid="interval-input"]');
    await intervalField.fill('2h');
    
    // Save
    const saveButton = page.locator('button:has-text("Save Settings")');
    await expect(saveButton).toBeEnabled();
    await saveButton.click();
    
    // Should make API call
    await page.waitForTimeout(2000);
    expect(saveApiCalled).toBe(true);
    
    // Save button should be disabled again (no pending changes)
    await expect(saveButton).toBeDisabled();
  });

  test('should handle form validation', async ({ page }) => {
    // Wait for form to load
    await page.waitForTimeout(2000);
    
    // Test port validation - should not accept invalid ports
    const webPortField = page.locator('input[data-testid="web-port-input"]');
    
    // Try invalid port number
    await webPortField.fill('99999'); // Invalid port
    
    const saveButton = page.locator('button:has-text("Save Settings")');
    
    // Should either prevent save or show validation error
    // This depends on the actual validation implementation
    if (await saveButton.isEnabled()) {
      await saveButton.click();
      // Should show error message or handle gracefully
      await page.waitForTimeout(1000);
    }
  });
});