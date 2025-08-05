# Plan 10: TDD Frontend Data Display Fixes

## Overview
This plan uses Test-Driven Development (TDD) methodology to systematically fix frontend data display issues discovered through Playwright E2E tests. Following Red-Green-Refactor cycle to ensure robust fixes.

## Status: 🔄 In Progress

### Completed ✅
- [x] **Setup Playwright E2E Testing Infrastructure**
  - Created comprehensive test suites for all pages
  - Configured ES module support and test runners
  - Added test scripts to package.json

- [x] **Dashboard Page Fixes** (5/6 tests passing)
  - Fixed API data structure parsing in `ApiService.request()`
  - Enhanced backend `/api/status` endpoint with complete system metrics
  - Added `GetUptime()` method to SystemMonitor with start time tracking
  - Added data-testid attributes for test accessibility
  - **Result**: Status cards now display actual data (uptime, running scripts, total scripts)

### In Progress 🔄

#### **Scripts Page Data Display** (Priority: High)
**Current Issues Identified:**
- Script cards showing incomplete information (empty names/paths)
- Only showing "Interval: s" and "Status: Disabled" 
- Missing script details from API responses

**TDD Test Cases:**
- [ ] Should display script names correctly
- [ ] Should display script paths correctly  
- [ ] Should show proper interval formatting (not just "s")
- [ ] Should display actual enable/disable status
- [ ] Should have functional action buttons

**Expected Fixes:**
- [ ] Add data-testid attributes to Scripts.vue component
- [ ] Fix API data extraction in useScripts composable
- [ ] Verify backend `/api/scripts` response format
- [ ] Fix interval display formatting in frontend

#### **Logs Page JavaScript Errors** (Priority: High)
**Current Issues Identified:**
- `TypeError: v.value.forEach is not a function` JavaScript errors
- Page stuck in permanent "Loading logs..." state
- Disabled Refresh and Clear Logs buttons
- Logs not displaying properly

**TDD Test Cases:**
- [ ] Should not have JavaScript console errors
- [ ] Should load logs successfully (not stuck loading)
- [ ] Should have functional filter controls
- [ ] Should have enabled action buttons
- [ ] Should display logs with proper formatting
- [ ] Should show log statistics correctly

**Expected Fixes:**
- [ ] Fix reactive data structure in useLogs composable
- [ ] Ensure logs API returns array format
- [ ] Add proper error handling for empty logs
- [ ] Add data-testid attributes to Logs.vue component

#### **Settings Page JavaScript Errors** (Priority: Medium)
**Current Issues Identified:**
- `TypeError: v.value.forEach is not a function` JavaScript errors  
- Configuration data may not be loading properly
- Form validation and state management issues

**TDD Test Cases:**
- [ ] Should not have JavaScript console errors
- [ ] Should load configuration from API successfully
- [ ] Should display all form fields with values
- [ ] Should enable save button when changes are made
- [ ] Should reset form correctly
- [ ] Should display system information correctly

**Expected Fixes:**
- [ ] Fix reactive data structure in Settings.vue
- [ ] Verify `/api/config` response parsing
- [ ] Add proper form validation
- [ ] Add data-testid attributes for testing

### Testing Strategy

#### TDD Workflow
1. **Red Phase**: Write failing tests that define expected behavior
2. **Green Phase**: Write minimal code to make tests pass
3. **Refactor Phase**: Clean up code while maintaining test passes

#### Test Categories
- **Unit Tests**: Composables and services (Vitest)
- **Integration Tests**: API communication (Vitest)  
- **E2E Tests**: Full user workflows (Playwright)

#### Success Criteria
- All Playwright E2E tests passing (100% pass rate)
- No JavaScript console errors on any page
- All data displays correctly from backend APIs
- Functional buttons and controls on all pages

### Implementation Notes

#### API Response Format
All backend APIs follow this structure:
```json
{
  "success": true,
  "data": {...},
  "error": "optional error message"
}
```

#### Frontend Data Flow
1. Vue components use composables for data management
2. Composables call ApiService methods
3. ApiService extracts `data` field from API responses
4. Reactive refs update Vue components automatically

#### Key Files Modified
- `web/frontend/src/services/api.ts` - API response parsing
- `web/server.go` - Backend status endpoint enhancement  
- `service/monitor.go` - System monitor uptime tracking
- `web/frontend/src/views/Dashboard.vue` - Data-testid attributes

### Next Steps
1. **Fix Scripts Page** - Add data-testid attributes and fix data display
2. **Fix Logs Page** - Resolve JavaScript errors and loading issues
3. **Fix Settings Page** - Fix form data loading and validation
4. **Run Full Test Suite** - Ensure all E2E tests pass
5. **Performance Testing** - Verify API response times
6. **Documentation Update** - Update component documentation

### Testing Commands
```bash
# Run all E2E tests
pnpm run test:e2e

# Run specific page tests
pnpm run test:e2e tests/e2e/specs/dashboard.spec.js
pnpm run test:e2e tests/e2e/specs/scripts.spec.js
pnpm run test:e2e tests/e2e/specs/logs.spec.js
pnpm run test:e2e tests/e2e/specs/settings.spec.js

# Debug tests with UI
pnpm run test:e2e:debug

# Run unit tests
pnpm test
```

### Acceptance Criteria
- [ ] All Playwright E2E tests passing (target: 100%)
- [ ] Zero JavaScript console errors across all pages
- [ ] All data fields displaying actual values (not empty/undefined)
- [ ] All buttons and controls functional
- [ ] Proper error handling for API failures
- [ ] Responsive design maintained across all fixes

### Risk Mitigation
- Maintain backward compatibility with existing APIs
- Preserve existing component styling and UX
- Ensure fixes don't break other functionality
- Add comprehensive error handling
- Document all changes for future maintenance

---

**Plan Status**: Active Development
**Started**: 2025-08-06
**Target Completion**: TBD based on complexity
**Dependencies**: Playwright testing infrastructure (completed)