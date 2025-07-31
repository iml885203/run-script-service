// Main Application Logic
class App {
    constructor() {
        this.currentTab = 'dashboard';
        this.scripts = [];
        this.logs = [];
        this.refreshInterval = null;
        this.autoRefreshEnabled = true;
        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.switchTab('dashboard');
        await this.loadInitialData();
        this.startAutoRefresh();
    }

    setupEventListeners() {
        // Tab navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const tab = link.dataset.tab;
                this.switchTab(tab);
            });
        });

        // Add script modal
        const addScriptBtn = document.getElementById('add-script-btn');
        if (addScriptBtn) {
            addScriptBtn.addEventListener('click', () => {
                components.openModal('add-script-modal');
            });
        }

        // Modal close buttons
        document.querySelectorAll('.modal-close').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const modal = e.target.closest('.modal');
                if (modal) {
                    components.closeModal(modal.id);
                }
            });
        });

        // Close modal on backdrop click
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    components.closeModal(modal.id);
                }
            });
        });

        // Add script form
        const addScriptForm = document.getElementById('add-script-form');
        if (addScriptForm) {
            addScriptForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleAddScript();
            });
        }

        // Cancel add script
        const cancelBtn = document.getElementById('cancel-add-script');
        if (cancelBtn) {
            cancelBtn.addEventListener('click', () => {
                components.closeModal('add-script-modal');
                components.resetForm('add-script-form');
            });
        }

        // Script filter
        const scriptFilter = document.getElementById('script-filter');
        if (scriptFilter) {
            scriptFilter.addEventListener('change', () => {
                this.loadLogs();
            });
        }

        // Refresh logs button
        const refreshLogsBtn = document.getElementById('refresh-logs-btn');
        if (refreshLogsBtn) {
            refreshLogsBtn.addEventListener('click', () => {
                this.loadLogs();
            });
        }

        // Clear logs button
        const clearLogsBtn = document.getElementById('clear-logs-btn');
        if (clearLogsBtn) {
            clearLogsBtn.addEventListener('click', () => {
                this.handleClearLogs();
            });
        }

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                // Close any open modals
                document.querySelectorAll('.modal.active').forEach(modal => {
                    components.closeModal(modal.id);
                });
            }
        });
    }

    switchTab(tabName) {
        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.toggle('active', link.dataset.tab === tabName);
        });

        // Update tab content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `${tabName}-tab`);
        });

        this.currentTab = tabName;

        // Load tab-specific data
        switch (tabName) {
            case 'dashboard':
                this.loadDashboard();
                break;
            case 'scripts':
                this.loadScripts();
                break;
            case 'logs':
                this.loadLogs();
                break;
            case 'settings':
                this.loadSettings();
                break;
        }
    }

    async loadInitialData() {
        try {
            await Promise.all([
                this.loadScripts(),
                this.loadSystemStatus()
            ]);
        } catch (error) {
            console.error('Failed to load initial data:', error);
            components.showNotification('Failed to load initial data', 'error');
        }
    }

    async loadSystemStatus() {
        try {
            const status = await apiClient.getStatus();
            this.systemStatus = status;
            components.updateDashboardMetrics(this.scripts, this.systemStatus);
        } catch (error) {
            console.error('Failed to load system status:', error);
        }
    }

    async loadScripts() {
        try {
            this.scripts = await apiClient.getScripts();
            components.renderScriptTable(this.scripts);
            components.updateScriptFilter(this.scripts);
            components.updateDashboardMetrics(this.scripts, this.systemStatus);
        } catch (error) {
            console.error('Failed to load scripts:', error);
            components.showNotification('Failed to load scripts', 'error');
        }
    }

    async loadLogs() {
        try {
            const scriptFilter = document.getElementById('script-filter');
            const selectedScript = scriptFilter ? scriptFilter.value : '';

            const options = {};
            if (selectedScript) {
                options.script = selectedScript;
            }
            options.limit = 50; // Limit to last 50 entries

            this.logs = await apiClient.getLogs(options);
            components.renderLogs(this.logs);
        } catch (error) {
            console.error('Failed to load logs:', error);
            components.showNotification('Failed to load logs', 'error');
        }
    }

    async loadDashboard() {
        // Dashboard loads data from other methods
        components.updateDashboardMetrics(this.scripts, this.systemStatus);
    }

    async loadSettings() {
        try {
            const config = await apiClient.getConfig();
            this.displaySystemConfig(config);
        } catch (error) {
            console.error('Failed to load settings:', error);
            components.showNotification('Failed to load settings', 'error');
        }
    }

    displaySystemConfig(config) {
        const configElement = document.getElementById('system-config');
        if (configElement && config) {
            configElement.innerHTML = `
                <pre>${JSON.stringify(config, null, 2)}</pre>
            `;
        }
    }

    async handleAddScript() {
        try {
            const formData = components.getFormData('add-script-form');
            if (!formData) return;

            await apiClient.addScript(formData);

            components.showNotification('Script added successfully', 'success');
            components.closeModal('add-script-modal');
            components.resetForm('add-script-form');

            await this.loadScripts();
        } catch (error) {
            console.error('Failed to add script:', error);
            components.showNotification(`Failed to add script: ${error.message}`, 'error');
        }
    }

    async runScript(scriptName) {
        try {
            await apiClient.runScript(scriptName);
            components.showNotification(`Script '${scriptName}' executed successfully`, 'success');

            // Refresh logs if we're on the logs tab
            if (this.currentTab === 'logs') {
                setTimeout(() => this.loadLogs(), 1000);
            }
        } catch (error) {
            console.error('Failed to run script:', error);
            components.showNotification(`Failed to run script '${scriptName}': ${error.message}`, 'error');
        }
    }

    async enableScript(scriptName) {
        try {
            await apiClient.enableScript(scriptName);
            components.showNotification(`Script '${scriptName}' enabled`, 'success');
            await this.loadScripts();
        } catch (error) {
            console.error('Failed to enable script:', error);
            components.showNotification(`Failed to enable script '${scriptName}': ${error.message}`, 'error');
        }
    }

    async disableScript(scriptName) {
        try {
            await apiClient.disableScript(scriptName);
            components.showNotification(`Script '${scriptName}' disabled`, 'success');
            await this.loadScripts();
        } catch (error) {
            console.error('Failed to disable script:', error);
            components.showNotification(`Failed to disable script '${scriptName}': ${error.message}`, 'error');
        }
    }

    async deleteScript(scriptName) {
        if (!confirm(`Are you sure you want to delete script '${scriptName}'?`)) {
            return;
        }

        try {
            await apiClient.deleteScript(scriptName);
            components.showNotification(`Script '${scriptName}' deleted`, 'success');
            await this.loadScripts();
        } catch (error) {
            console.error('Failed to delete script:', error);
            components.showNotification(`Failed to delete script '${scriptName}': ${error.message}`, 'error');
        }
    }

    async handleClearLogs() {
        const scriptFilter = document.getElementById('script-filter');
        const selectedScript = scriptFilter ? scriptFilter.value : '';

        if (selectedScript) {
            if (!confirm(`Clear logs for script '${selectedScript}'?`)) return;

            try {
                await apiClient.clearScriptLogs(selectedScript);
                components.showNotification(`Logs cleared for script '${selectedScript}'`, 'success');
                await this.loadLogs();
            } catch (error) {
                console.error('Failed to clear script logs:', error);
                components.showNotification(`Failed to clear logs: ${error.message}`, 'error');
            }
        } else {
            components.showNotification('Please select a script to clear logs for', 'warning');
        }
    }

    startAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }

        this.refreshInterval = setInterval(() => {
            if (this.autoRefreshEnabled) {
                // Refresh current tab data
                switch (this.currentTab) {
                    case 'dashboard':
                        this.loadSystemStatus();
                        break;
                    case 'scripts':
                        this.loadScripts();
                        break;
                    case 'logs':
                        this.loadLogs();
                        break;
                }
            }
        }, 30000); // Refresh every 30 seconds
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    toggleAutoRefresh() {
        this.autoRefreshEnabled = !this.autoRefreshEnabled;
        if (this.autoRefreshEnabled) {
            this.startAutoRefresh();
        }
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new App();
});
